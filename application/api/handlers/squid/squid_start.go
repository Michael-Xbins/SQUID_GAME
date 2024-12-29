package squid

import (
	"application/api/presenter"
	"application/api/presenter/squid"
	"application/mongodb"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"application/session"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var currentTime int64

const scaleFactor = int64(1000) // 千分比,保持赔率的精度

const allDead = "重新计算赔率,所有人死亡"
const noDead = "重新计算赔率,没有人死亡"
const reOdds = "重新计算赔率,原赔率 < 1.01"

func StartSquidGame() {
	utils.LoopSafeGo(startSquid)
}

func startSquid() {
	var init bool
	for {
		hasUser := false
		timestamp, v, err := redis.BPop(redis.SquidPrefix)
		if err != nil {
			log.Error(err)
			continue
		}
		currentTime = timestamp
		game := &squid.Game{}
		if e := mongodb.Find(context.Background(), game, 0); e != nil {
			log.Error(e)
			continue
		}
		if !init {
			if e := squidOpen(game, timestamp); e != nil {
				log.Error("鱿鱼游戏开盘失败", e)
			}
			init = true
		}
		session.S2AllMessage(&pb.ClientResponse{
			Type: pb.MessageType_SquidPriceChangeNotify_,
			Message: &pb.ClientResponse_SquidPriceChangeNotify{
				SquidPriceChangeNotify: &pb.SquidPriceChangeNotify{
					Num:       v,
					Timestamp: timestamp},
			},
		})

		if game.State == 0 {
			if timestamp >= game.CloseTime {
				deadTrackTime, nextSecondTransPrice, err := redis.BPop(redis.SquidPrefix) //死亡赛道时间比封盘时间晚1秒
				game.DeadTrackTime = deadTrackTime
				if err != nil {
					log.Error(err)
					continue
				}
				if e := squidClose(game, nextSecondTransPrice); e != nil {
					log.Error("鱿鱼游戏封盘失败", e)
				}
			}
		}

		if game.State == 1 {
			if timestamp >= game.ResultTime {
				if e := squidRoundEnd(timestamp, game); e != nil {
					log.Error("鱿鱼游戏结算失败", e)
				}
				players, _ := redis.GetSquidPlayers()
				if len(players) > 0 {
					hasUser = true
				}

				// 停服处理, 不允许下单
				//if viper.GetBool("common.stopGame") {
				//	if e := mongodb.Update(context.Background(), game, nil); e != nil {
				//		log.Error(e)
				//	}
				//	//go mongodb.Check(mongodb.SquidType) // 临时, 核对总账单, 总出口 == 总入口
				//	log.Info("正在停服, 木头人停止服务")
				//	for viper.GetBool("common.stopGame") {
				//		time.Sleep(3 * time.Second)
				//	}
				//}

			}
		}

		if game.State == 2 && timestamp >= game.EndTime {
			if e := squidOpen(game, timestamp); e != nil {
				log.Error("鱿鱼游戏开盘失败", e)
			}
		}
		if e := mongodb.Update(context.Background(), game, nil); e != nil {
			log.Error(e)
			continue
		}
		if hasUser {
			go mongodb.Check(mongodb.SquidType) // 临时, 核对总账单, 总出口 == 总入口
		}
	}
}

func squidOpen(game *squid.Game, timestamp int64) error {
	log.Debug("鱿鱼游戏开始")
	game.RoundNum++
	game.State = 0
	game.DeadTrack = -1
	openTime := int64(utils.LubanTables.TBWood.Get("open_time").NumInt)
	closeTime := int64(utils.LubanTables.TBWood.Get("close_time").NumInt)
	settlementTime := int64(utils.LubanTables.TBWood.Get("settlement_time").NumInt)
	game.StartTime = timestamp
	game.CloseTime = timestamp + openTime*1000
	game.DeadTrackTime = -1
	game.ResultTime = timestamp + (openTime+closeTime)*1000
	game.EndTime = timestamp + (openTime+closeTime+settlementTime)*1000
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_SquidStageNotify_,
		Message: &pb.ClientResponse_SquidStageNotify{SquidStageNotify: &pb.SquidStageNotify{
			Stage:     game.State,
			Countdown: game.SquidCountdown(currentTime),
			Timestamp: currentTime,
		}},
	})

	hasRobot := viper.GetBool("common.has_robot")
	if hasRobot {
		if err := RobotOrder(game); err != nil {
			log.Error("RobotOrder error: ", err)
		}
	}
	return nil
}

func squidClose(game *squid.Game, transPrice float64) error {
	log.Debug("鱿鱼游戏封盘")
	deadTrack, _, err := calculateDeadTrack(transPrice)
	if err != nil {
		return err
	}
	game.State = 1
	game.TransPrice = transPrice
	game.DeadTrack = deadTrack
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_SquidStageNotify_,
		Message: &pb.ClientResponse_SquidStageNotify{SquidStageNotify: &pb.SquidStageNotify{
			Stage:      game.State,
			Countdown:  game.SquidCountdown(currentTime),
			Timestamp:  currentTime,
			TransPrice: transPrice,
			DeadTrack:  deadTrack,
		}},
	})

	//TODO:遍历下注玩家，时间到了没有下注视为退赛; 更新玩家的流水额
	players, _ := redis.GetSquidPlayers()
	if len(players) > 0 {
		for _, player := range players {
			userInfo := &presenter.UserInfo{}
			if err := mongodb.Find(context.Background(), userInfo, player); err != nil {
				log.Error(err)
			}
			if userInfo.Squid.BetPrices == 0 {
				if err := redis.RemoveSquidPlayer(player); err != nil {
					log.Errorf("user: %s, err: %s", player, err)
				}
				mongodb.SquidReset(userInfo) // 玩家退赛重置
			} else {
				// 更新玩家的流水额(下注的钱)
				mongodb.AddSquidTurnOver(userInfo, userInfo.Squid.BetPrices)
			}
			if e := mongodb.Update(context.Background(), userInfo, nil); e != nil {
				log.Errorf("user: %s, err: %s", userInfo.Account, e)
				continue
			}
		}
	}
	return nil
}

func calculateDeadTrack(value float64) (int32, int32, error) {
	// 将浮点数转换为字符串
	valueStr := fmt.Sprintf("%.2f", value) // 保留两位小数
	// 移除小数点
	valueStr = strings.Replace(valueStr, ".", "", 1)
	// 计算所有数字的和
	sum := 0
	for _, char := range valueStr {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			fmt.Println("Error converting character to integer:", err)
			return -1, -1, err
		}
		sum += digit
	}
	// 对4取余，得到死亡赛道编号
	deadTrack := int32((sum % 4) + 1) // 赛道编号从1开始
	return deadTrack, int32(sum), nil
}

func squidRoundEnd(timestamp int64, game *squid.Game) error {
	log.Debug("鱿鱼游戏结束, 进行结算")

	game.State = 2
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_SquidStageNotify_,
		Message: &pb.ClientResponse_SquidStageNotify{SquidStageNotify: &pb.SquidStageNotify{
			Stage:      game.State,
			Countdown:  game.SquidCountdown(timestamp),
			Timestamp:  timestamp,
			TransPrice: game.TransPrice,
			DeadTrack:  game.DeadTrack,
		}},
	})

	//TODO:获取本次死亡赛道,计算赔率、结算,更新userinfo,更新squidFund
	squidFund := &squid.Fund{}
	if err := mongodb.Find(context.Background(), squidFund, 0); err != nil {
		log.Error(err)
		return err
	}
	players, _ := redis.GetSquidPlayers()
	jackPotPlayers := make([]string, 0)
	// 计算并存储所有玩家已完成轮数的jackpot权重比例
	allPlayerWeights, _, _ := calculateAllPlayerWeights(players)

	// 理论赔付总额 (总下注额 减去 4笔抽水)
	totalRoundBets := make([]int64, squid.TotalRounds) // 初始存储 每个轮次的总注额
	// 实际赔付总额
	winRoundBets := make([]int64, squid.TotalRounds) // 每个轮次实际的赔付总额, 初始存储 获胜机器人每个轮次的总下注额

	// 本期玩家每个轮次的总注额
	playerBets := make([]int64, squid.TotalRounds)
	// 本期机器人每个轮次的总注额
	robotBets := make([]int64, squid.TotalRounds)

	for roundId := 1; roundId <= squid.TotalRounds; roundId++ {
		totalBet, deadTrackBet, totalUserBet, userDeadTrackBet, totalRobotBet, robotDeadTrackBet := calculateBets(int32(roundId), game.DeadTrack)
		log.Debugf("轮次%d, 所有赛道总投注额%d,死亡赛道总注额%d, 玩家所有赛道总投注额%d,玩家死亡赛道总投注额%d, 机器人所有赛道总投注额%d,机器人死亡赛道总投注额%d", roundId, totalBet, deadTrackBet, totalUserBet, userDeadTrackBet, totalRobotBet, robotDeadTrackBet)
		totalRoundBets[roundId-1] = totalBet
		playerBets[roundId-1] = totalUserBet
		robotBets[roundId-1] = totalRobotBet
	}

	for _, player := range players {
		userInfo := &presenter.UserInfo{}
		if err := mongodb.Find(context.Background(), userInfo, player); err != nil {
			log.Errorf("player:%s, squidRoundEnd error:%s", player, err)
			continue
		}
		//结算
		settlement(userInfo, squidFund, game, allPlayerWeights, totalRoundBets, winRoundBets)
		if userInfo.Squid.CanJackpot {
			jackPotPlayers = append(jackPotPlayers, userInfo.Account) //jackpot获奖用户
		}
	}

	if e := mongodb.Update(context.Background(), squidFund, nil); e != nil {
		log.Error("Error updating squid fund:", e)
	}

	//TODO:处理jackpot,首通奖
	jackpotPayout := int64(0)
	if len(jackPotPlayers) != 0 {
		processFirstPass(jackPotPlayers)
		jackpotPayout = processJackpot(jackPotPlayers)
	}

	if err := mongodb.Find(context.Background(), squidFund, 0); err != nil {
		log.Error("Error retrieving squid fund:", err)
	}
	//TODO:通知client当前每一轮的赔率
	rounds := make([]*pb.Round, squid.TotalRounds)
	for roundId := int32(1); roundId <= squid.TotalRounds; roundId++ {
		totalBet, deadTrackBet, _, _, _, _ := calculateBets(roundId, game.DeadTrack)
		odds, _, _, _ := calRoundOdds(roundId, game.DeadTrack)
		realOdds := float64(odds) / float64(scaleFactor)
		rounds[roundId-1] = &pb.Round{
			RoundId:  roundId,
			Odds:     realOdds,
			TotalBet: totalBet,
			DeadBet:  deadTrackBet,
		}

		if totalRoundBets[roundId-1] > winRoundBets[roundId-1] {
			log.Debugf("轮次%d, 赔率:%v, 本轮(理论)赔付总额: %d, 本轮(实际)赔付总额: %d, 精度损失: %d", roundId, realOdds, totalRoundBets[roundId-1], winRoundBets[roundId-1], totalRoundBets[roundId-1]-winRoundBets[roundId-1])
			squidFund.RobotPool += totalRoundBets[roundId-1] - winRoundBets[roundId-1] // 精度损耗补充到机器人库存
		}
	}
	if err := mongodb.Update(context.Background(), squidFund, nil); err != nil {
		log.Error("Error updating squid fund:", err)
	}

	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_SquidRoundNotify_,
		Message: &pb.ClientResponse_SquidRoundNotify{SquidRoundNotify: &pb.SquidRoundNotify{
			Rounds:       rounds,
			TotalJackpot: squidFund.Jackpot,
		}},
	})

	//TODO:redis存储每次的开奖数据
	_, btcSum, _ := calculateDeadTrack(game.TransPrice)
	newSquidFund := &squid.Fund{}
	if err := mongodb.Find(context.Background(), newSquidFund, 0); err != nil {
		log.Error("Error retrieving squid fund:", err)
	}
	playerRate := float64(0)
	robotRate := float64(0)
	if squidFund.PlayerAccAmount.AccBet > 0 {
		playerRate = float64(squidFund.PlayerAccAmount.AccPayout) / float64(squidFund.PlayerAccAmount.AccBet)
	}
	if squidFund.RobotAccAmount.AccBet > 0 {
		robotRate = float64(squidFund.RobotAccAmount.AccPayout) / float64(squidFund.RobotAccAmount.AccBet)
	}
	data := squid.Data{
		Timestamp: time.UnixMilli(game.DeadTrackTime).Format("2006-01-02 15:04:05"),
		BtcPrice:  game.TransPrice, // BTC价格
		BtcSum:    btcSum,          // BTC价之和
		DeadTrack: game.DeadTrack,  // 死亡赛道

		Odds1:     rounds[0].Odds,     // 第1轮赔率
		TotalBet1: rounds[0].TotalBet, // 第1轮总注额
		DeadBet1:  rounds[0].DeadBet,  // 第1轮死亡注额

		Odds2:     rounds[1].Odds,     // 第2轮赔率
		TotalBet2: rounds[1].TotalBet, // 第2轮总注额
		DeadBet2:  rounds[1].DeadBet,  // 第2轮死亡注额

		Odds3:     rounds[2].Odds,     // 第3轮赔率
		TotalBet3: rounds[2].TotalBet, // 第3轮总注额
		DeadBet3:  rounds[2].DeadBet,  // 第3轮死亡注额

		Odds4:     rounds[3].Odds,     // 第4轮赔率
		TotalBet4: rounds[3].TotalBet, // 第4轮总注额
		DeadBet4:  rounds[3].DeadBet,  // 第4轮死亡注额

		Odds5:     rounds[4].Odds,     // 第5轮赔率
		TotalBet5: rounds[4].TotalBet, // 第5轮总注额
		DeadBet5:  rounds[4].DeadBet,  // 第5轮死亡注额

		Odds6:     rounds[5].Odds,     // 第6轮赔率
		TotalBet6: rounds[5].TotalBet, // 第6轮总注额
		DeadBet6:  rounds[5].DeadBet,  // 第6轮死亡注额

		Odds7:     rounds[6].Odds,     // 第7轮赔率
		TotalBet7: rounds[6].TotalBet, // 第7轮总注额
		DeadBet7:  rounds[6].DeadBet,  // 第7轮死亡注额

		HouseCut:  newSquidFund.HouseCut, // 庄家抽水余额
		Jackpot:   newSquidFund.Jackpot,  // Jackpot余额
		RobotPool: squidFund.RobotPool,   // 机器人库存

		PlayerAccBet:    squidFund.PlayerAccAmount.AccBet,    // 玩家累计投注额
		PlayerAccPayout: squidFund.PlayerAccAmount.AccPayout, // 玩家累计赔付额
		PlayerRate:      playerRate,                          // 玩家反水比例
		RobotAccBet:     squidFund.RobotAccAmount.AccBet,     // 机器人累计投注额
		RobotAccPayout:  squidFund.RobotAccAmount.AccPayout,  // 机器人累计赔付额
		RobotRate:       robotRate,                           // 机器人反水比例
	}
	if err := redis.AddSquidData(game.CloseTime, data); err != nil {
		log.Errorf("CloseTime: %d, AddSquidData error: %v", game.CloseTime, err)
	}

	robotBalance, _ := mongodb.SumRobotBalances(context.Background())
	log.Infof("日志核对, 木头人第%d期, 本期死亡赛道:%d,transPrice:%f, 本期玩家每轮注额%d, 本期机器人每轮注额%d, 本期每轮赔付总额(包含机器人):%d, 机器人总余额:%d, jackpot领取人数:%d,jackpot领取总额:%d",
		game.RoundNum, game.DeadTrack, game.TransPrice, playerBets, robotBets, winRoundBets, robotBalance, len(jackPotPlayers), jackpotPayout)

	//TODO:重置GlobalSquidRound
	mongodb.ResetAllGlobalSquidRound()
	//for i := 1; i <= squid.TotalRounds; i++ {
	//	globalSquidRound := &squid.GlobalRound{}
	//	if err := mongodb.Find(context.Background(), globalSquidRound, i); err != nil {
	//		log.Error("Error finding round:", err)
	//		continue
	//	}
	//	mongodb.ResetGlobalSquidRound(globalSquidRound)
	//	if err := mongodb.Update(context.Background(), globalSquidRound, nil); err != nil {
	//		log.Error("Error updating round:", err)
	//		continue
	//	}
	//}
	return nil
}

// 获取本轮所有赛道的总注额和死亡赛道的注额
func calculateBets(roundId int32, deadTrack int32) (int64, int64, int64, int64, int64, int64) {
	globalSquidRound := &squid.GlobalRound{}
	if err := mongodb.Find(context.Background(), globalSquidRound, roundId); err != nil {
		log.Error("Error finding round:", err)
	}
	tracks := []*squid.Track{&globalSquidRound.Track1, &globalSquidRound.Track2, &globalSquidRound.Track3, &globalSquidRound.Track4}
	totalBet := int64(0)     // 本轮4个赛道的总注额之和
	deadTrackBet := int64(0) // 死亡赛道的总注额

	userTotalBet := int64(0)     // 玩家总注额之和
	userDeadTrackBet := int64(0) // 死亡玩家总注额

	robotTotalBet := int64(0)     // 机器人总注额之和
	robotDeadTrackBet := int64(0) // 死亡机器人总注额

	for i, track := range tracks {
		totalBet += track.TotalBetPrices + track.RobotTotalBetPrices
		userTotalBet += track.TotalBetPrices
		robotTotalBet += track.RobotTotalBetPrices
		if i+1 == int(deadTrack) {
			deadTrackBet = track.TotalBetPrices + track.RobotTotalBetPrices
			userDeadTrackBet = track.TotalBetPrices
			robotDeadTrackBet = track.RobotTotalBetPrices
		}
	}
	return totalBet, deadTrackBet, userTotalBet, userDeadTrackBet, robotTotalBet, robotDeadTrackBet
}

// 计算赔率
func calRoundOdds(roundId int32, deadTrack int32) (int64, int64, string, bool) {
	// 获取本轮所有赛道的总注额和死亡赛道的注额
	totalBet, deadTrackBet, _, _, _, _ := calculateBets(roundId, deadTrack)
	if totalBet == 0 {
		//log.Warnf("第%d轮 无人下注: ", roundId)
		return 0, 0, "", false
	}
	// 计算赔率
	pumpProfit := int64(utils.LubanTables.TBWood.Get("pump_profit").NumInt)   //千分比
	pumpJackpot := int64(utils.LubanTables.TBWood.Get("pump_jackpot").NumInt) //千分比
	pumpDay := int64(utils.LubanTables.TBWood.Get("pump_day").NumInt)         //千分比
	pumpActing := int64(utils.LubanTables.TBWood.Get("pump_acting").NumInt)   //千分比
	x := totalBet * (scaleFactor - (pumpProfit + pumpJackpot + pumpDay + pumpActing))
	y := totalBet - deadTrackBet
	var reOddsType string
	odds := int64(0)
	oldOdds := odds
	isRecalculated := false
	if totalBet > deadTrackBet {
		odds = x / y
		oldOdds = odds
	}
	if odds < 1010 {
		// 重新计算赔率
		isRecalculated = true
		if deadTrackBet == 0 {
			// 没有人死亡：所有人不抽水,拿回投注额
			odds = scaleFactor * 1
			reOddsType = noDead
		} else if totalBet <= deadTrackBet {
			// 所有人死亡, 按5笔抽水, 其中可赔付抽水给庄家
			odds = 0
			reOddsType = allDead
		} else {
			y := totalBet - deadTrackBet
			x := deadTrackBet*(scaleFactor-(pumpProfit+pumpJackpot+pumpDay+pumpActing)) + scaleFactor*y
			odds = x / y
			reOddsType = reOdds
		}
	}
	return odds, oldOdds, reOddsType, isRecalculated
}

// 获取抽水详情
func getPumpDetails(userRoundBet int64) *presenter.PumpDetails {
	scaleFactor := int64(1000)                                                // 千分比,保持赔率的精度
	pumpProfit := int64(utils.LubanTables.TBWood.Get("pump_profit").NumInt)   //千分比
	pumpJackpot := int64(utils.LubanTables.TBWood.Get("pump_jackpot").NumInt) //千分比
	pumpDay := int64(utils.LubanTables.TBWood.Get("pump_day").NumInt)         //千分比
	pumpActing := int64(utils.LubanTables.TBWood.Get("pump_acting").NumInt)   //千分比
	acting0 := int64(utils.LubanTables.TBWood.Get("acting0").NumInt)          //千分比
	acting1 := int64(utils.LubanTables.TBWood.Get("acting1").NumInt)          //千分比
	agent := userRoundBet * pumpActing / scaleFactor
	upLineAgent := userRoundBet * pumpActing * acting0 / scaleFactor / scaleFactor
	upUpLineAgent := agent - upLineAgent

	details := &presenter.PumpDetails{
		PumpProfit:            pumpProfit,
		PumpJackpot:           pumpJackpot,
		PumpDay:               pumpDay,
		PumpActing:            pumpActing,
		Acting0:               acting0,
		Acting1:               acting1,
		HouseCut:              userRoundBet * pumpProfit / scaleFactor,
		JackpotContribution:   userRoundBet * pumpJackpot / scaleFactor,
		FirstPassContribution: userRoundBet * pumpDay / scaleFactor,
		AgentContribution:     agent,
		UpLineContribution:    upLineAgent,
		UpUpLineContribution:  upUpLineAgent,
	}
	return details
}

func settlement(userInfo *presenter.UserInfo, squidFund *squid.Fund, game *squid.Game, allPlayerWeights map[string]int64, totalRoundBets []int64, winRoundBets []int64) {
	currentRoundId := userInfo.Squid.RoundId
	// 结算后websocket通知字段
	realOdds := float64(0)
	bonus := int64(0)
	jackpotWeightRate := float64(0)
	selectTrack := userInfo.Squid.Track      // 玩家本轮所选赛道
	userRoundBet := userInfo.Squid.BetPrices // 玩家本轮注额

	if !userInfo.IsRobot {
		log.Debugf("玩家%s, 选择赛道: %d, 下注金额: %d, 当前轮次: %d",
			userInfo.Account, selectTrack, userRoundBet, userInfo.Squid.RoundId)
	}

	// 计算玩家已完成轮数的jackpot权重比例
	totalWeight := int64(0)
	for _, weight := range allPlayerWeights {
		totalWeight += weight
	}
	weight := allPlayerWeights[userInfo.Account]
	jackpotWeightRate = float64(weight) / float64(totalWeight)

	// 获取本轮所有赛道的总注额和死亡赛道的注额
	totalBet, deadTrackBet, _, _, _, _ := calculateBets(currentRoundId, game.DeadTrack)
	if totalBet == 0 {
		log.Infof("轮次%d 所有赛道均没有额度: ", currentRoundId)
		return
	}

	// 获取本轮抽水详情
	pumpDetails := getPumpDetails(userRoundBet)

	// 计算赔率
	odds, oldOdds, reOddsType, isRecalculated := calRoundOdds(currentRoundId, game.DeadTrack)
	if isRecalculated {
		log.Debugf("轮次%d 需要重新计算赔率, 原赔率: %v, 原因: %v, 新赔率: %v", currentRoundId, oldOdds, reOddsType, odds)
	} else {
		log.Debugf("计算轮次%d 赔率为: %v", currentRoundId, odds)
	}

	//规则：
	//赔率 >= 1.01时,所有人的钱分5笔抽水,并且赢的人分bonus
	//赔率 < 1.01时：
	//  (1)没有人死亡： 所有人不抽水,拿回投注额;
	//  (2)所有人死亡:  按5笔抽水, 其中可赔付抽水给庄家;
	//  (3)有死亡有存活: 重新计算赔率,仅死亡赛道的玩家进行5笔抽水(庄家,agent佣金,jackpot,每日首通,赔付),其中的赔付抽水,赔偿给其余玩家作为奖金.

	if !isRecalculated {
		realOdds = float64(odds) / float64(scaleFactor) // 真实赔率
		log.Debugf("轮次%d odds>=1.01, account: %v, odds: %v, upline: %v, oldHouseCut: %v, oldJackpot: %v, 是否机器人: %v, 本次抽水 庄家(houseCut)/机器人(robotPool): %v, jackpot: %v, 首通: %v, 代理: %v,代理上线:%v,代理上上线:%v",
			currentRoundId, userInfo.Account, realOdds, userInfo.UpLine, squidFund.HouseCut, squidFund.Jackpot, userInfo.IsRobot, pumpDetails.HouseCut, pumpDetails.JackpotContribution, pumpDetails.FirstPassContribution, pumpDetails.AgentContribution, pumpDetails.UpLineContribution, pumpDetails.UpUpLineContribution)
		// 更新抽水: houseCut,jackpot,firstPass,agent
		mongodb.AddSquidHouseCut(userInfo.IsRobot, squidFund, pumpDetails.HouseCut)
		mongodb.AddSquidJackpot(squidFund, pumpDetails.JackpotContribution)
		mongodb.AddFirstPass(userInfo, squidFund, pumpDetails.FirstPassContribution)
		mongodb.AddSquidAgent(userInfo, squidFund, pumpDetails)
		if userInfo.Squid.Track == game.DeadTrack {
			mongodb.SquidReset(userInfo) // 玩家死亡重置
		} else {
			bonus = odds * userRoundBet / scaleFactor
			mongodb.AddAmount(userInfo, bonus) // 更新玩家余额
			mongodb.SquidNextRound(userInfo)   // 完成,进入下一轮
		}
		totalRoundBets[currentRoundId-1] = totalRoundBets[currentRoundId-1] - pumpDetails.HouseCut - pumpDetails.JackpotContribution - pumpDetails.FirstPassContribution - pumpDetails.AgentContribution

	} else {
		if deadTrackBet == 0 {
			// 没有人死亡：所有人不抽水,拿回投注额
			log.Debugf("轮次%d, odds<1.01, 没有人死亡", currentRoundId)
			realOdds = 1
			log.Debugf("轮次%d, odds<1.01, account: %v, newOdds: %v, upline: %v, oldHouseCut: %v, oldJackpot: %v, 是否机器人: %v, 本次抽水 庄家(houseCut)/机器人(robotPool): %v, jackpot: %v, 首通: %v, 代理: %v,代理上线:%v,代理上上线:%v",
				currentRoundId, userInfo.Account, realOdds, userInfo.UpLine, squidFund.HouseCut, squidFund.Jackpot, userInfo.IsRobot, 0, 0, 0, 0, 0, 0)
			bonus = userRoundBet
			mongodb.AddAmount(userInfo, userRoundBet)
			mongodb.SquidNextRound(userInfo)
		} else if totalBet <= deadTrackBet {
			// 所有人死亡, 按5笔抽水, 其中可赔付抽水给庄家
			log.Debugf("轮次%d, odds<1.01, 所有人死亡", currentRoundId)
			newHouseCut := userRoundBet - (pumpDetails.JackpotContribution + pumpDetails.FirstPassContribution + pumpDetails.AgentContribution)
			log.Debugf("轮次%d, odds<1.01, account: %v, newOdds: %v, upline: %v, oldHouseCut: %v, oldJackpot: %v, 是否机器人: %v, 本次抽水 庄家(houseCut)/机器人(robotPool): %v, jackpot: %v, 首通: %v, 代理: %v,代理上线:%v,代理上上线:%v",
				currentRoundId, userInfo.Account, realOdds, userInfo.UpLine, squidFund.HouseCut, squidFund.Jackpot, userInfo.IsRobot, newHouseCut, pumpDetails.JackpotContribution, pumpDetails.FirstPassContribution, pumpDetails.AgentContribution, pumpDetails.UpLineContribution, pumpDetails.UpUpLineContribution)
			// 更新抽水: 庄家,jackpot,firstPass,agent
			mongodb.AddSquidHouseCut(userInfo.IsRobot, squidFund, newHouseCut)
			mongodb.AddSquidJackpot(squidFund, pumpDetails.JackpotContribution)
			mongodb.AddFirstPass(userInfo, squidFund, pumpDetails.FirstPassContribution)
			mongodb.AddSquidAgent(userInfo, squidFund, pumpDetails)
			mongodb.SquidReset(userInfo) // 玩家死亡重置
			totalRoundBets[currentRoundId-1] = 0
		} else {
			realOdds = float64(odds) / float64(scaleFactor) // 真实赔率
			if userInfo.Squid.Track == game.DeadTrack {
				// 仅死亡赛道的玩家进行5笔抽水
				log.Debugf("轮次%d odds<1.01, account: %v, newOdds: %v, upline: %v, oldHouseCut: %v, oldJackpot: %v, 是否机器人: %v, 本次抽水 庄家(houseCut)/机器人(robotPool): %v, jackpot: %v, 首通: %v, 代理: %v,代理上线:%v,代理上上线:%v",
					currentRoundId, userInfo.Account, realOdds, userInfo.UpLine, squidFund.HouseCut, squidFund.Jackpot, userInfo.IsRobot, pumpDetails.HouseCut, pumpDetails.JackpotContribution, pumpDetails.FirstPassContribution, pumpDetails.AgentContribution, pumpDetails.UpLineContribution, pumpDetails.UpUpLineContribution)
				// 更新抽水: 庄家,jackpot,firstPass,agent
				mongodb.AddSquidHouseCut(userInfo.IsRobot, squidFund, pumpDetails.HouseCut)
				mongodb.AddSquidJackpot(squidFund, pumpDetails.JackpotContribution)
				mongodb.AddFirstPass(userInfo, squidFund, pumpDetails.FirstPassContribution)
				mongodb.AddSquidAgent(userInfo, squidFund, pumpDetails)
				mongodb.SquidReset(userInfo) // 玩家死亡重置
				totalRoundBets[currentRoundId-1] = totalRoundBets[currentRoundId-1] - pumpDetails.HouseCut - pumpDetails.JackpotContribution - pumpDetails.FirstPassContribution - pumpDetails.AgentContribution

			} else {
				log.Debugf("轮次%d odds<1.01, account: %v, odds: %v, upline: %v, oldHouseCut: %v, oldJackpot: %v,  是否机器人: %v, 本次抽水 庄家(houseCut)/机器人(robotPool): %v, jackpot: %v, 首通: %v, 代理: %v,代理上线:%v,代理上上线:%v",
					currentRoundId, userInfo.Account, realOdds, userInfo.UpLine, squidFund.HouseCut, squidFund.Jackpot, userInfo.IsRobot, 0, 0, 0, 0, 0, 0)
				bonus = userRoundBet * odds / scaleFactor
				mongodb.AddAmount(userInfo, bonus) // 更新玩家余额
				mongodb.SquidNextRound(userInfo)   // 完成,进入下一轮
			}
		}
	}
	if userInfo.IsRobot {
		squidFund.RobotAccAmount.AccBet += userRoundBet
		squidFund.RobotAccAmount.AccPayout += bonus
	} else {
		squidFund.PlayerAccAmount.AccBet += userRoundBet
		squidFund.PlayerAccAmount.AccPayout += bonus
	}

	winRoundBets[currentRoundId-1] += bonus
	log.Debugf("user: %s, 轮次%d, weight: %v, totalWeight: %v, jackpotWeightRate: %v, bonus: %v, newHouseCut: %v, newJackpot: %v, 玩家累计投注总额: %v, 玩家累计赔付总额: %v, 机器人累计投注总额: %v, 机器人累计赔付总额: %v\n",
		userInfo.Account, currentRoundId, weight, totalWeight, jackpotWeightRate, bonus, squidFund.HouseCut, squidFund.Jackpot, squidFund.PlayerAccAmount.AccBet, squidFund.PlayerAccAmount.AccPayout, squidFund.RobotAccAmount.AccBet, squidFund.RobotAccAmount.AccPayout)

	// 如上userinfo更新信息写入db
	if err := mongodb.Update(context.Background(), userInfo, nil); err != nil {
		log.Errorf("userInfo: %v, Error updating user: %s", userInfo, err)
	}

	//记录订单
	isWin := false
	if selectTrack != game.DeadTrack {
		isWin = true
	}
	_, btcSum, _ := calculateDeadTrack(game.TransPrice)
	orderInfo := &squid.Order{
		OrderId:    uuid.New().String(),
		IsRobot:    userInfo.IsRobot,
		Account:    userInfo.Account,
		RoundNum:   game.RoundNum,
		Track:      selectTrack,
		BetPrices:  userRoundBet,
		TransPrice: game.TransPrice,
		BtcSum:     btcSum,
		Odds:       realOdds,
		IsWin:      isWin,
		Bonus:      bonus,
		Timestamp:  time.Now().UnixMilli(),
	}
	if err := mongodb.Insert(context.Background(), orderInfo); err != nil {
		log.Error("insert order error: ", err)
	}

	session.S2CMessage(userInfo.Account, &pb.ClientResponse{
		Type: pb.MessageType_SquidSettleDoneNotify_,
		Message: &pb.ClientResponse_SquidSettleDoneNotify{SquidSettleDoneNotify: &pb.SquidSettleDoneNotify{
			DeadTrack:         game.DeadTrack,
			TransPrice:        game.TransPrice,
			Bets:              userRoundBet,
			Odds:              realOdds,
			Bonus:             bonus,
			JackpotWeightRate: jackpotWeightRate,
			NextRoundId:       userInfo.Squid.RoundId,
			Balance:           userInfo.Balance,
		}},
	})
	if userInfo.Squid.RoundId < squid.TotalRounds {
		session.S2CMessage(userInfo.Account, &pb.ClientResponse{
			Type: pb.MessageType_SquidJackpotNotify_,
			Message: &pb.ClientResponse_SquidJackpotNotify{SquidJackpotNotify: &pb.SquidJackpotNotify{
				FirstPass: userInfo.Squid.FirstPass.Pool,
				Jackpot:   0,
			}},
		})
	}
}

func processJackpot(jackPotPlayers []string) int64 {
	if len(jackPotPlayers) == 0 {
		return 0
	}
	// 获取jackpot总额
	squidFund := &squid.Fund{}
	if err := mongodb.Find(context.Background(), squidFund, 0); err != nil {
		log.Error("Error retrieving squid fund:", err)
		return 0
	}

	// 计算本次发放的总奖金
	jackpotMin := int(utils.LubanTables.TBWood.Get("jackpot_min").NumInt)
	jackpotMax := int(utils.LubanTables.TBWood.Get("jackpot_max").NumInt)
	jackpotPercentage := rand.Intn(jackpotMax-jackpotMin+1) + jackpotMin
	totalJackpot := squidFund.Jackpot * int64(jackpotPercentage) / 100

	// 计算所有玩家的个人权重, 并获取玩家信息
	allPlayerWeights, userInfos, err := calculateAllPlayerWeights(jackPotPlayers)
	if err != nil {
		log.Error("Error calculating player weights or retrieving user info:", err)
		return 0
	}

	// 计算所有玩家的总权重
	totalWeight := int64(0)
	for _, weight := range allPlayerWeights {
		totalWeight += weight
	}

	// 分发jackpot奖金, 批量更新userinfo
	realTotalJackpot := int64(0)
	var userInfosToUpdate []*presenter.UserInfo
	for player, userInfo := range userInfos {
		weight := allPlayerWeights[player]
		userJackpot := totalJackpot * weight / totalWeight
		log.Debugf("processJackpot, user: %v, Jackpot总奖池:%d,本次比例%f,本次奖池%v, 玩家权重:%v,所有玩家总权重:%v, jackpot奖金: %v",
			player, squidFund.Jackpot, float64(jackpotPercentage)/100, totalJackpot, weight, totalWeight, userJackpot)
		mongodb.AddAmount(userInfo, userJackpot)
		realTotalJackpot += userJackpot //考虑精度损失,使用实际扣除款
		userInfosToUpdate = append(userInfosToUpdate, userInfo)
		mongodb.SquidReset(userInfo) // 玩家jackpot后重置
		session.S2CMessage(userInfo.Account, &pb.ClientResponse{
			Type: pb.MessageType_SquidJackpotNotify_,
			Message: &pb.ClientResponse_SquidJackpotNotify{SquidJackpotNotify: &pb.SquidJackpotNotify{
				FirstPass: userInfo.Squid.FirstPass.Pool,
				Jackpot:   userJackpot,
			}},
		})
	}
	if err := mongodb.BulkUpdateUserInfos(context.Background(), userInfosToUpdate); err != nil {
		log.Error("Error updating user balances:", err)
	}

	// 更新jackpot总额
	mongodb.DecrSquidJackpot(squidFund, realTotalJackpot)
	if err := mongodb.Update(context.Background(), squidFund, nil); err != nil {
		log.Error("Error updating squid fund:", err)
	}
	return realTotalJackpot
}

func calculateAllPlayerWeights(jackPotPlayers []string) (map[string]int64, map[string]*presenter.UserInfo, error) {
	allPlayerWeights := make(map[string]int64)
	userInfos := make(map[string]*presenter.UserInfo)
	for _, player := range jackPotPlayers {
		userInfo := &presenter.UserInfo{}
		if err := mongodb.Find(context.Background(), userInfo, player); err != nil {
			log.Error("Error retrieving user info for player:", player, err)
			continue
		}
		userInfos[player] = userInfo
		weight := calculateUserWeight(userInfo)
		allPlayerWeights[player] = weight
	}
	return allPlayerWeights, userInfos, nil
}

func calculateUserWeight(userInfo *presenter.UserInfo) int64 {
	weights := []int64{
		int64(utils.LubanTables.TBWood.Get("money_power1").NumInt),
		int64(utils.LubanTables.TBWood.Get("money_power2").NumInt),
		int64(utils.LubanTables.TBWood.Get("money_power3").NumInt),
		int64(utils.LubanTables.TBWood.Get("money_power4").NumInt),
		int64(utils.LubanTables.TBWood.Get("money_power5").NumInt),
		int64(utils.LubanTables.TBWood.Get("money_power6").NumInt),
		int64(utils.LubanTables.TBWood.Get("money_power7").NumInt),
	}
	// 计算玩家已完成轮数的权重总和
	totalWeight := int64(0)
	for i, bet := range userInfo.Squid.BetPricesPerRound {
		if i < len(weights) {
			totalWeight += bet * weights[i]
			//log.Debugf("calculateUserWeight, user: %v, %d轮下注总数:%v, weight%d: %v, 累计至本轮权重: %v", userInfo.Account, i+1, bet, i+1, weights[i], totalWeight)
		}
	}
	return totalWeight
}

func processFirstPass(jackPotPlayers []string) {
	if len(jackPotPlayers) == 0 {
		return
	}
	var userInfosToUpdate []*presenter.UserInfo
	for _, player := range jackPotPlayers {
		userInfo := &presenter.UserInfo{}
		if err := mongodb.Find(context.Background(), userInfo, player); err != nil {
			log.Error("Error retrieving user info for first pass process:", err)
			continue
		}
		//每天首次领取：更新余额, 重置个人首通奖励池
		mongodb.SquidDailyFirstPass(userInfo)
		userInfosToUpdate = append(userInfosToUpdate, userInfo)
	}
	if len(userInfosToUpdate) > 0 {
		if err := mongodb.BulkUpdateUserInfos(context.Background(), userInfosToUpdate); err != nil {
			log.Error("Error updating user balances:", err)
		}
	}
}
