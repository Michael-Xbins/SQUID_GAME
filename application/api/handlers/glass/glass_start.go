package glass

import (
	"application/api/presenter"
	"application/api/presenter/glass"
	"application/mongodb"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"application/session"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/spf13/viper"
	"strconv"
	"time"
	"unicode"
)

var currentTime int64

func StartGlassGame() {
	utils.LoopSafeGo(startGlass)
}

func startGlass() {
	var init bool
	for {
		hasUser := false //是否有玩家下注
		timestamp := time.Now().UnixMilli()
		currentTime = timestamp
		game := &glass.Game{}
		if e := mongodb.Find(context.Background(), game, 0); e != nil {
			log.Error(e)
			continue
		}
		if !init {
			if e := glassOpen(game, timestamp); e != nil {
				log.Error("玻璃桥游戏开盘失败", e)
			}
			init = true
		}

		if game.State == 0 {
			if timestamp >= game.CloseTime {
				if e := glassClose(game); e != nil {
					log.Error("玻璃桥游戏封盘失败", e)
				}
			}
		}

		if game.State == 1 {
			if timestamp >= game.ResultTime {
				players, _ := redis.GetGlassPlayers()
				if len(players) > 0 {
					hasUser = true
				}
				if e := glassRoundEnd(timestamp, game); e != nil {
					log.Error("玻璃桥游戏结算失败", e)
				}

				// 停服处理, 不允许下单
				if viper.GetBool("common.stopGame") {
					if e := mongodb.Update(context.Background(), game, nil); e != nil {
						log.Error(e)
					}
					log.Info("正在停服, 玻璃桥停止服务")
					for viper.GetBool("common.stopGame") {
						time.Sleep(3 * time.Second)
					}
				}

			}
		}

		if game.State == 2 && timestamp >= game.EndTime {
			if e := glassOpen(game, timestamp); e != nil {
				log.Error("玻璃桥游戏开盘失败", e)
			}
		}
		if e := mongodb.Update(context.Background(), game, nil); e != nil {
			log.Error(e)
			continue
		}
		if hasUser {
			//go mongodb.Check(mongodb.GlassType) // 临时, 核对总账单, 总出口 == 总入口
		}
	}
}

func glassOpen(game *glass.Game, timestamp int64) error {
	log.Debug("玻璃桥游戏开始")
	game.RoundNum++
	game.State = 0
	openTime := int64(utils.LubanTables.TBGlass.Get("open_time").NumInt)
	closeTime := int64(utils.LubanTables.TBGlass.Get("close_time").NumInt)
	settlementTime := int64(utils.LubanTables.TBGlass.Get("settlement_time").NumInt)
	game.StartTime = timestamp
	game.CloseTime = timestamp + openTime*1000
	game.ResultTime = timestamp + (openTime+closeTime)*1000
	game.EndTime = timestamp + (openTime+closeTime+settlementTime)*1000
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_GlassStageNotify_,
		Message: &pb.ClientResponse_GlassStageNotify{GlassStageNotify: &pb.GlassStageNotify{
			Stage:     game.State,
			RoundNum:  game.RoundNum,
			Countdown: game.GlassCountdown(currentTime),
			Timestamp: currentTime,
		}},
	})
	return nil
}

func glassClose(game *glass.Game) error {
	log.Debug("玻璃桥游戏封盘")
	var transHash string
	var err error

	if viper.GetString("common.env") == utils.Produce {
		//transHash, err = wallet.Transfer()
	} else {
		transHash, err = calculateTransHash()
		if err != nil {
			return err
		}
	}

	game.State = 1
	game.TransHash = transHash
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_GlassStageNotify_,
		Message: &pb.ClientResponse_GlassStageNotify{GlassStageNotify: &pb.GlassStageNotify{
			Stage:     game.State,
			RoundNum:  game.RoundNum,
			Countdown: game.GlassCountdown(currentTime),
			Timestamp: currentTime,
			TransHash: transHash,
		}},
	})

	winNum, hashNum, err := calWinNum(transHash)
	if err != nil {
		log.Error("calWinNum error: ", err)
		return err
	}
	log.Infof("日志核对, 玻璃桥第%d期, transHash: %v, winNum: %v, hashNum: %v", game.RoundNum, transHash, winNum, hashNum)
	lotteryInfo := &glass.Lottery{
		RoundNum:  game.RoundNum,
		Hash1:     int32(hashNum[0]),
		Hash2:     int32(hashNum[1]),
		Hash3:     int32(hashNum[2]),
		Track1:    int32(winNum[0]),
		Track2:    int32(winNum[1]),
		Track3:    int32(winNum[2]),
		Timestamp: time.Now().UnixMilli(),
	}
	if err := mongodb.Insert(context.Background(), lotteryInfo); err != nil {
		log.Error("insert lottery error: ", err)
		return err
	}
	return nil
}
func calculateTransHash() (string, error) {
	//return "0x0ae5d981e054be5c65f22ab50c679f75d9273f5a51774e95c2454877068f7fd0", nil
	// 生成15字节的随机数，因为每个字节可以表示为两个16进制字符
	bytes := make([]byte, 15)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err // 返回错误
	}
	// 将字节转换为16进制字符串
	hash := hex.EncodeToString(bytes)
	return "0x" + hash, nil
}

func glassRoundEnd(timestamp int64, game *glass.Game) error {
	log.Debug("玻璃桥游戏结束, 进行结算")
	game.State = 2
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_GlassStageNotify_,
		Message: &pb.ClientResponse_GlassStageNotify{GlassStageNotify: &pb.GlassStageNotify{
			Stage:     game.State,
			RoundNum:  game.RoundNum,
			Countdown: game.GlassCountdown(timestamp),
			Timestamp: timestamp,
			TransHash: game.TransHash,
		}},
	})

	//TODO:结算,更新userinfo,更新glassFund
	glassFund := &glass.Fund{}
	if err := mongodb.Find(context.Background(), glassFund, 0); err != nil {
		log.Error(err)
		return err
	}

	var userInfosToUpdate []*presenter.UserInfo //批量更新userinfo
	var OrdersToUpdate []*glass.Order           //批量更新order
	players, _ := redis.GetGlassPlayers()
	winNum, _, err := calWinNum(game.TransHash)
	if err != nil {
		log.Error("calWinNum error: ", err)
	}
	for _, player := range players {
		userInfo := &presenter.UserInfo{}
		if err := mongodb.Find(context.Background(), userInfo, player); err != nil {
			log.Errorf("player:%s, glassRoundEnd error:%s", player, err)
			continue
		}
		//结算
		winningOrders, err := settlement(userInfo, glassFund, game.RoundNum, winNum)
		if err != nil {
			log.Error("Settlement error: ", err)
			continue
		}

		userInfosToUpdate = append(userInfosToUpdate, userInfo)
		OrdersToUpdate = append(OrdersToUpdate, winningOrders...)
		if err := redis.RemoveGlassPlayer(player); err != nil {
			log.Errorf("user: %s, err: %s", player, err)
		}
	}
	if len(userInfosToUpdate) > 0 {
		if err := mongodb.BulkUpdateUserInfos(context.Background(), userInfosToUpdate); err != nil {
			log.Errorf("userInfosToUpdate:%v, Error updating user balances:%s", userInfosToUpdate, err)
		}
		if e := mongodb.Update(context.Background(), glassFund, nil); e != nil {
			log.Error("Error updating glass fund:", e)
		}
	}
	if len(OrdersToUpdate) > 0 {
		if err := mongodb.BulkUpdateOrderInfos(context.Background(), OrdersToUpdate); err != nil {
			log.Errorf("Error updating order infos: %s", err)
		}
	}
	return nil
}
func settlement(userInfo *presenter.UserInfo, glassFund *glass.Fund, RoundNum int64, winNum []int) ([]*glass.Order, error) {
	game := &glass.Game{}
	if e := mongodb.Find(context.Background(), game, 0); e != nil {
		log.Error(e)
		return nil, e
	}
	curOrders, err := mongodb.FindGlassUserOrders(context.Background(), userInfo.Account, RoundNum)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var winningOrders []*glass.Order //批量更新订单
	orders := make([]*pb.Order, len(curOrders))
	for i, order := range curOrders {
		match := true
		bonus := float64(0)
		if len(order.Track1) != 0 {
			match = match && (strconv.Itoa(winNum[0]) == order.Track1)
		}
		if len(order.Track2) != 0 {
			match = match && (strconv.Itoa(winNum[1]) == order.Track2)
		}
		if len(order.Track3) != 0 {
			match = match && (strconv.Itoa(winNum[2]) == order.Track3)
		}

		if match {
			betType := order.BetType
			bonus = float64(utils.LubanTables.TBGlass.Get(betType).NumInt) * order.Odds
			glassFund.AvailableFund -= int64(bonus)
			if order.TrackType == 1 {
				glassFund.AccAmount1.AccPayout += int64(bonus)
			} else if order.TrackType == 2 {
				glassFund.AccAmount2.AccPayout += int64(bonus)
			} else {
				glassFund.AccAmount3.AccPayout += int64(bonus)
			}
			mongodb.AddAmount(userInfo, int64(bonus))
			order.IsWin = true
			order.Bonus = int64(bonus)
			winningOrders = append(winningOrders, order)
		}
		//通知client
		orders[i] = &pb.Order{
			OrderId: order.OrderId,
			IsWin:   order.IsWin,
			Odds:    order.Odds,
			Bonus:   bonus,
		}
	}

	session.S2CMessage(userInfo.Account, &pb.ClientResponse{
		Type: pb.MessageType_GlassSettleDoneNotify_,
		Message: &pb.ClientResponse_GlassSettleDoneNotify{GlassSettleDoneNotify: &pb.GlassSettleDoneNotify{
			RoundNum: RoundNum,
			Orders:   orders,
			Balance:  userInfo.Balance,
		}},
	})
	return winningOrders, nil
}

func calWinNum(transHash string) ([]int, []int, error) {
	// 找到最后六个数字
	var digits string
	for i := len(transHash) - 1; i >= 0 && len(digits) < 6; i-- {
		if unicode.IsDigit(rune(transHash[i])) {
			digits = string(transHash[i]) + digits
		}
	}
	if len(digits) < 6 {
		return nil, nil, fmt.Errorf("not enough digits found in the string")
	}

	// 分为三组
	group1, _ := strconv.Atoi(digits[0:2])
	group2, _ := strconv.Atoi(digits[2:4])
	group3, _ := strconv.Atoi(digits[4:6])
	result1 := group1 % 3
	result2 := group2 % 4
	result3 := group3 % 3

	return []int{result1, result2, result3}, []int{group1, group2, group3}, nil
}
