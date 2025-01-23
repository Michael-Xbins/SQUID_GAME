package ladder

import (
	"application/api/presenter"
	"application/api/presenter/ladder"
	"application/mongodb"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"application/session"
	"application/wallet"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

var currentTime int64

func StartLadderGame() {
	utils.LoopSafeGo(startLadder)
}

var funChan = make(chan func(game *ladder.Game), 256)

func startLadder() {
	var init bool
	ticker := time.NewTicker(1 * time.Second)
	game := &ladder.Game{}
	if e := mongodb.Find(context.Background(), game, 0); e != nil {
		log.Error(e)
		time.Sleep(3 * time.Second)
		return
	}
	for {
		select {
		case f := <-funChan:
			f(game)
		case <-ticker.C:
		}
		hasUser := false //是否有玩家下注
		timestamp := time.Now().UnixMilli()
		currentTime = timestamp
		if !init {
			if e := ladderOpen(game, timestamp); e != nil {
				log.Error("梯子游戏开盘失败", e)
			}
			init = true
		}

		if game.State == 0 {
			if timestamp >= game.CloseTime {
				if e := ladderClose(game); e != nil {
					log.Error("梯子游戏封盘失败", e)
				}
			}
		}

		if game.State == 1 {
			if timestamp >= game.ResultTime {
				players, _ := redis.GetLadderPlayers(game.RoundNum)
				if len(players) > 0 {
					hasUser = true
				}
				if e := ladderRoundEnd(timestamp, game); e != nil {
					log.Error("梯子游戏结算失败", e)
				}
			}
		}

		if game.State == 2 && timestamp >= game.EndTime {
			if e := ladderOpen(game, timestamp); e != nil {
				log.Error("梯子游戏开盘失败", e)
			}
		}
		if e := mongodb.Update(context.Background(), game, nil); e != nil {
			log.Error(e)
			continue
		}
		if hasUser {
			//go mongodb.Check(mongodb.LadderType) // 临时, 核对总账单, 总出口 == 总入口
		}
	}
}

func ladderOpen(game *ladder.Game, timestamp int64) error {
	log.Debug("梯子游戏开始")
	game.RoundNum++
	game.State = 0

	openTime := int64(utils.LubanTables.TBLadder.Get("open_time").NumInt)
	closeTime := int64(utils.LubanTables.TBLadder.Get("close_time").NumInt)
	settlementTime := int64(utils.LubanTables.TBLadder.Get("settlement_time").NumInt)
	game.StartTime = timestamp
	game.CloseTime = timestamp + openTime*1000
	game.ResultTime = timestamp + (openTime+closeTime)*1000
	game.EndTime = timestamp + (openTime+closeTime+settlementTime)*1000
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_LadderStageNotify_,
		Message: &pb.ClientResponse_LadderStageNotify{LadderStageNotify: &pb.LadderStageNotify{
			Stage:     game.State,
			RoundNum:  game.RoundNum,
			Countdown: game.LadderCountdown(currentTime),
			Timestamp: currentTime,
		}},
	})
	return nil
}

func ladderClose(game *ladder.Game) error {
	log.Debug("梯子游戏封盘")
	var err error
	game.State = 1
	bets, _ := redis.GetLadderBets(context.Background(), game.RoundNum)
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_LadderStageNotify_,
		Message: &pb.ClientResponse_LadderStageNotify{LadderStageNotify: &pb.LadderStageNotify{
			Stage:     game.State,
			RoundNum:  game.RoundNum,
			Countdown: game.LadderCountdown(currentTime),
			Timestamp: currentTime,
			Bets:      bets,
		}},
	})

	roundNum := game.RoundNum
	utils.SafeGo(func() {
		var transHash string
		if viper.GetString("common.env") == utils.Produce {
			transHash, err = wallet.Transfer(roundNum)
		} else {
			transHash, err = calculateTransHash()
		}
		if err != nil {
			log.Error("Error during transfer or hash calculation: ", err)
			return
		}
		select {
		case funChan <- func(game *ladder.Game) {
			err := processLadderRoundEnd(roundNum, transHash)
			if err != nil {
				log.Error("processLadderRoundEnd error: ", err)
			}
		}:
		default:
			fmt.Println("队列阻塞") // channel最大容量256
		}
	})
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

func ladderRoundEnd(timestamp int64, game *ladder.Game) error {
	log.Debug("梯子游戏结束, 进行结算")
	game.State = 2
	bets, _ := redis.GetLadderBets(context.Background(), game.RoundNum)
	lottery := &ladder.Lottery{}
	if err := mongodb.Find(context.Background(), lottery, game.RoundNum); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warnf("梯子交易hash还没有生成出来, 期数:%d", game.RoundNum)
		} else {
			log.Error(err)
		}
	}
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_LadderStageNotify_,
		Message: &pb.ClientResponse_LadderStageNotify{LadderStageNotify: &pb.LadderStageNotify{
			Stage:     game.State,
			RoundNum:  game.RoundNum,
			Countdown: game.LadderCountdown(timestamp),
			Timestamp: timestamp,
			TransHash: lottery.Hash,
			Bets:      bets,
		}},
	})
	return nil
}

func processLadderRoundEnd(roundNum int64, transHash string) error {
	//TODO:结算,更新userinfo,更新ladderFund
	ladderFund := &ladder.Fund{}
	if err := mongodb.Find(context.Background(), ladderFund, 0); err != nil {
		log.Error(err)
		return err
	}

	var userInfosToUpdate []*presenter.UserInfo //批量更新userinfo
	var OrdersToUpdate []*ladder.Order          //批量更新order
	players, _ := redis.GetLadderPlayers(roundNum)
	winNum, hashNum, err := calWinNum(transHash)
	if err != nil {
		log.Error("calWinNum error: ", err)
		return err
	}
	var totalPayout1, totalPayout2 int64 //本期直注赔付额,二串一赔付额

	for _, player := range players {
		userInfo := &presenter.UserInfo{}
		if err := mongodb.Find(context.Background(), userInfo, player); err != nil {
			log.Errorf("player:%s, ladderRoundEnd error:%s", player, err)
			continue
		}
		//结算
		winningOrders, payout1, payout2, err := settlement(userInfo, ladderFund, roundNum, winNum)
		if err != nil {
			log.Error("Settlement error: ", err)
			continue
		}
		userInfosToUpdate = append(userInfosToUpdate, userInfo)
		OrdersToUpdate = append(OrdersToUpdate, winningOrders...)
		if err := redis.RemoveLadderPlayer(player, roundNum); err != nil {
			log.Errorf("user: %s, err: %s", player, err)
		}

		totalPayout1 += payout1
		totalPayout2 += payout2
	}
	if len(userInfosToUpdate) > 0 {
		if err := mongodb.BulkUpdateUserInfos(context.Background(), userInfosToUpdate); err != nil {
			log.Errorf("userInfosToUpdate:%v, Error updating user balances:%s", userInfosToUpdate, err)
			return err
		}
		if e := mongodb.Update(context.Background(), ladderFund, nil); e != nil {
			log.Error("Error updating ladder fund:", e)
			return e
		}
	}
	if len(OrdersToUpdate) > 0 {
		if err := mongodb.BulkUpdateLadderOrderInfos(context.Background(), OrdersToUpdate); err != nil {
			log.Errorf("Error updating ladder order infos: %s", err)
			return err
		}
	}

	sumCategory1, sumCategory2, _ := mongodb.SumTotalPricesByCategory(context.Background(), roundNum)
	backWater1 := float64(0)
	backWater2 := float64(0)
	if ladderFund.AccAmount1.AccBet > 0 {
		backWater1 = float64(ladderFund.AccAmount1.AccPayout) / float64(ladderFund.AccAmount1.AccBet)
	}
	if ladderFund.AccAmount2.AccBet > 0 {
		backWater2 = float64(ladderFund.AccAmount2.AccPayout) / float64(ladderFund.AccAmount2.AccBet)
	}
	if err := redis.RemoveLadderBets(context.Background(), roundNum); err != nil {
		log.Error("RemoveLadderBets error: ", err)
	}

	lotteryInfo := &ladder.Lottery{
		RoundNum:  roundNum,
		Hash:      transHash,
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

	log.Infof("日志核对, 梯子第%d期, winNum: %v,transHash: %v, 本期直注总额:%d,直注赔付额:%d,直注历史反水比例:%f, 本期二串一总额%d,二串一赔付额:%d,二串一历史反水比例:%f", roundNum, winNum, transHash, sumCategory1, totalPayout1, backWater1, sumCategory2, totalPayout2, backWater2)
	return nil
}

func settlement(userInfo *presenter.UserInfo, ladderFund *ladder.Fund, RoundNum int64, winNum []int) ([]*ladder.Order, int64, int64, error) {
	oldBalance := userInfo.Balance
	curOrders, err := mongodb.FindLadderUserOrders(context.Background(), userInfo.Account, RoundNum)
	if err != nil {
		log.Error(err)
		return nil, 0, 0, err
	}

	var winningOrders []*ladder.Order //批量更新订单
	orders := make([]*pb.Order, len(curOrders))
	betPrices := int64(0)
	var totalPayoutCategory1, totalPayoutCategory2 int64 // 玩家userInfo的直注和二串一赔付总额
	for i, order := range curOrders {
		match := true
		bonus := float64(0)
		betPrices += order.OrderPrice

		// 分解 BetId，可能包含多个赛道信息
		betId := order.BetId
		betId = strings.TrimPrefix(betId, "round_") // 移除前缀
		for i := 0; i < len(betId); i += 2 {
			track, _ := strconv.Atoi(betId[i : i+1])      // 赛道编号
			position, _ := strconv.Atoi(betId[i+1 : i+2]) // 位置编号
			// 检查是否匹配获胜位置
			if position != winNum[track-1] {
				match = false
				break
			}
		}

		if match {
			bonus = float64(order.OrderPrice) * order.Odds
			ladderFund.AvailableFund -= int64(bonus)
			if order.Category == 1 {
				ladderFund.AccAmount1.AccPayout += int64(bonus)
				totalPayoutCategory1 += int64(bonus)
			} else if order.Category == 2 {
				ladderFund.AccAmount2.AccPayout += int64(bonus)
				totalPayoutCategory2 += int64(bonus)
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

	// 更新每日任务进度
	mongodb.UpdateDailyTaskProgress(5, userInfo, 1)

	if userInfo.Balance-oldBalance-betPrices > 0 {
		log.InfoJson("金币入口", // coinFlow埋点
			zap.String("Account", userInfo.Account),
			zap.String("ActionType", log.Flow),
			zap.String("FlowType", log.CoinFlow),
			zap.String("From", log.FromLadderSettlement),
			zap.String("Flag", log.FlagIn),
			zap.Int64("RoundNum", RoundNum),
			zap.Int64("Amount", userInfo.Balance-oldBalance-betPrices),
			zap.Int64("Old", oldBalance+betPrices),
			zap.Int64("New", userInfo.Balance),
			zap.Int64("CreatedAt", time.Now().UnixMilli()),
		)
	} else if userInfo.Balance-oldBalance-betPrices < 0 {
		log.InfoJson("金币出口", // coinFlow埋点
			zap.String("Account", userInfo.Account),
			zap.String("ActionType", log.Flow),
			zap.String("FlowType", log.CoinFlow),
			zap.String("From", log.FromLadderSettlement),
			zap.String("Flag", log.FlagOut),
			zap.Int64("RoundNum", RoundNum),
			zap.Int64("Amount", oldBalance+betPrices-userInfo.Balance),
			zap.Int64("Old", oldBalance+betPrices),
			zap.Int64("New", userInfo.Balance),
			zap.Int64("CreatedAt", time.Now().UnixMilli()),
		)
	}

	session.S2CMessage(userInfo.Account, &pb.ClientResponse{
		Type: pb.MessageType_LadderSettleDoneNotify_,
		Message: &pb.ClientResponse_LadderSettleDoneNotify{LadderSettleDoneNotify: &pb.LadderSettleDoneNotify{
			RoundNum: RoundNum,
			Orders:   orders,
			Balance:  userInfo.Balance,
		}},
	})
	return winningOrders, totalPayoutCategory1, totalPayoutCategory2, nil
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
