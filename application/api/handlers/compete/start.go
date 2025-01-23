package compete

import (
	"application/api/presenter"
	"application/api/presenter/compete"
	"application/mongodb"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"application/session"
	"context"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"math"
	"time"
)

var currentTime int64

func StartCompeteGame() {
	utils.LoopSafeGo(EventLoop)
}

func EventLoop() {
	var init bool
	for {
		hasUser := false //是否有玩家下注
		timestamp, v, err := redis.BPop(redis.Prefix)
		if err != nil {
			log.Error(err)
			continue
		}
		currentTime = timestamp
		game := &compete.Game{}
		if e := mongodb.Find(context.Background(), game, 0); e != nil {
			log.Error(e)
			continue
		}
		if !init {
			if e := open(game, timestamp); e != nil {
				log.Error("拔河游戏开始失败", e)
			}
			init = true
		}
		round := game.CurRound
		closeTime := round.CloseTime
		resultTime := round.ResultTime
		endTime := round.EndTime
		if game.State != 2 {
			round.Nums = append(round.Nums, v)
			round.Sum += v
			session.S2AllMessage(&pb.ClientResponse{
				Type: pb.MessageType_NumberChangeNotify_,
				Message: &pb.ClientResponse_NumberChangeNotify{
					NumberChangeNotify: &pb.NumberChangeNotify{
						Num:       round.Sum,
						Timestamp: timestamp},
				},
			})
		}
		if game.State == 0 {
			if timestamp >= closeTime {
				if e := close(game); e != nil {
					log.Error("拔河游戏封盘失败", e)
				}
			}
		}
		if game.State == 1 {
			if timestamp >= resultTime {
				if e := roundEnd(timestamp, game); e != nil {
					log.Error("拔河游戏结算失败", e)
				}
				if len(game.CurRound.Orders) > 0 {
					hasUser = true
				}

				// 停服处理, 不允许下单
				if viper.GetBool("common.stopGame") {
					if e := mongodb.Update(context.Background(), game, nil); e != nil {
						log.Error(e)
					}
					log.Info("正在停服, 拔河停止服务")
					for viper.GetBool("common.stopGame") {
						time.Sleep(3 * time.Second)
					}
				}
			}
		}

		if game.State == 2 && timestamp >= endTime {
			if e := open(game, timestamp); e != nil {
				log.Error("拔河游戏开始失败", e)
			}
		}
		if e := mongodb.Update(context.Background(), game, nil); e != nil {
			log.Error(e)
			continue
		}
		if hasUser {
			//go mongodb.Check(mongodb.CompeteType) // 临时, 核对总账单, 总出口 == 总入口
		}
	}
}

func roundEnd(timestamp int64, game *compete.Game) error {
	log.Debug("拔河游戏结束")
	game.State = 2
	round := game.CurRound
	aOdds := int64(utils.LubanTables.TBCompete.Get("odds_a").NumInt)
	peaceOdds := int64(utils.LubanTables.TBCompete.Get("odds_peace").NumInt)
	bOdds := int64(utils.LubanTables.TBCompete.Get("odds_b").NumInt)

	num := round.Sum
	numA := int32(math.Floor(num*100)) % 10
	numB := int32(math.Floor(num*1000)) % 10
	var hit string
	var odds int64
	if numA < numB {
		hit = "b"
		odds = bOdds
	} else if numA > numB {
		hit = "a"
		odds = aOdds
	} else {
		hit = "peace"
		odds = peaceOdds
	}

	totalTakeAmount := int64(0)
	agentAmount := int64(0)
	takePercent := int64(utils.LubanTables.TBCompete.Get("pump_profit").NumInt) //千分比
	pumpActing := int64(utils.LubanTables.TBGlass.Get("pump_acting").NumInt)    //千分比
	acting0 := int64(utils.LubanTables.TBGlass.Get("acting0").NumInt)           //千分比

	lossAmount := int64(0)
	betTotalAmount := int64(0)
	for account, order := range round.Orders {
		winAmount := int64(0)
		betPrices := int64(0)
		betPricesA := int64(0)
		betPricesB := int64(0)
		betPricesPeace := int64(0)
		var pumpDetails presenter.PumpDetails
		if len(order.Amounts) > 0 {
			val := order.Amounts[hit]
			if val > 0 {
				winAmount = val * odds
				lossAmount += winAmount
			}
			for k, v := range order.Amounts {
				if v > 0 {
					betTotalAmount += v
					orderBonus := int64(0)
					if k == hit {
						orderBonus = v * odds
					}
					// 订单
					orderInfo := &compete.OrderInfo{
						OrderId:     uuid.New().String(),
						Account:     account,
						RoundNum:    game.RoundNum,
						TransAmount: num,
						Hit:         hit,
						Track:       k,
						IsWin:       k == hit,
						BetAmount:   v,
						Bonus:       orderBonus,
						Timestamp:   time.Now().UnixMilli(),
					}
					if k == "a" {
						orderInfo.Odds = aOdds
						game.AccAmountA.AccBet += v
						game.AccAmountA.AccPayout += orderBonus
						betPricesA += v
					} else if k == "b" {
						orderInfo.Odds = bOdds
						game.AccAmountB.AccBet += v
						game.AccAmountB.AccPayout += orderBonus
						betPricesB += v
					} else {
						orderInfo.Odds = peaceOdds
						game.AccAmountPeace.AccBet += v
						game.AccAmountPeace.AccPayout += orderBonus
						betPricesPeace += v
					}
					if err := mongodb.Insert(context.Background(), orderInfo); err != nil {
						log.Error("insert order error: ", err)
					}
				}
				betPrices += v
				totalTakeAmount += v * takePercent / 1000
				agentContribution := v * pumpActing / 1000
				upLineContribution := v * pumpActing * acting0 / 1000 / 1000
				upUpLineContribution := agentContribution - upLineContribution
				agentAmount += agentContribution

				pumpDetails.AgentContribution += agentContribution
				pumpDetails.UpLineContribution += upLineContribution
				pumpDetails.UpUpLineContribution += upUpLineContribution
			}
		}
		userInfo := &presenter.UserInfo{}
		if err := mongodb.Find(context.Background(), userInfo, account); err != nil {
			log.Error(err)
			continue
		}
		oldBalance := userInfo.Balance
		if winAmount > 0 {
			mongodb.AddAmount(userInfo, winAmount)
		}
		// 更新玩家的流水额(下注的钱)
		mongodb.AddCompeteTurnOver(userInfo, betPrices)
		// 更新代理抽水
		mongodb.AddCompeteAgent(userInfo, game, pumpDetails)

		//拔河本轮下注额(暂存), 用于核对账单, 结算/取消清零
		userInfo.CompeteLastBet = 0

		// 更新每日任务进度
		mongodb.UpdateDailyTaskProgress(4, userInfo, 1)

		if err := mongodb.Update(context.Background(), userInfo, nil); err != nil {
			log.Error(err)
			continue
		}
		session.S2CMessage(userInfo.Account, &pb.ClientResponse{
			Type: pb.MessageType_CompeteSettleDoneNotify_,
			Message: &pb.ClientResponse_CompeteSettleDoneNotify{CompeteSettleDoneNotify: &pb.CompeteSettleDoneNotify{
				Balance: userInfo.Balance,
			}},
		})

		log.Debugf("拔河订单：account: %v, 本轮a总注额%d,b总注额%d,peace总注额%d, hit: %v, odds: %v, lossAmount: %v, upline: %v, oldHouseCut: %v,  本次抽水包含 庄家: %v, 代理: %v,代理上线:%v,代理上上线:%v",
			account, betPricesA, betPricesB, betPricesPeace, hit, odds, lossAmount, userInfo.UpLine, game.TakeAmount, totalTakeAmount, pumpDetails.AgentContribution, pumpDetails.UpLineContribution, pumpDetails.UpUpLineContribution)

		if userInfo.Balance-oldBalance-betPrices > 0 {
			log.InfoJson("金币入口", // coinFlow埋点
				zap.String("Account", userInfo.Account),
				zap.String("ActionType", log.Flow),
				zap.String("FlowType", log.CoinFlow),
				zap.String("From", log.FromCompeteSettlement),
				zap.String("Flag", log.FlagIn),
				zap.Int64("RoundNum", game.RoundNum),
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
				zap.String("From", log.FromCompeteSettlement),
				zap.String("Flag", log.FlagOut),
				zap.Int64("RoundNum", game.RoundNum),
				zap.Int64("Amount", oldBalance+betPrices-userInfo.Balance),
				zap.Int64("Old", oldBalance+betPrices),
				zap.Int64("New", userInfo.Balance),
				zap.Int64("CreatedAt", time.Now().UnixMilli()),
			)
		}
	}

	game.Amount += betTotalAmount - totalTakeAmount - lossAmount - agentAmount
	game.TakeAmount += totalTakeAmount
	game.ResultHistoryList = append(game.ResultHistoryList, compete.ResultHistory{
		Hit:    hit,
		Round:  game.RoundNum,
		Volume: game.CurRound.Sum,
	})
	if len(game.ResultHistoryList) > 50 {
		game.ResultHistoryList = game.ResultHistoryList[1:]
	}
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_GameResultNotify_,
		Message: &pb.ClientResponse_GameResultNotify{
			GameResultNotify: &pb.GameResultNotify{
				WinColor: hit,
				NumA:     numA,
				NumB:     numB,
			}},
	})

	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_GameStageNotify_,
		Message: &pb.ClientResponse_GameStageNotify{
			GameStageNotify: &pb.GameStageNotify{
				Stage:     game.State,
				Countdown: game.Countdown(timestamp),
				Timestamp: timestamp,
			}},
	})

	backWater := float64(0)
	if game.AccAmountA.AccBet+game.AccAmountB.AccBet+game.AccAmountPeace.AccBet > 0 {
		backWater = float64(game.AccAmountA.AccPayout+game.AccAmountB.AccPayout+game.AccAmountPeace.AccPayout) / float64(game.AccAmountA.AccBet+game.AccAmountB.AccBet+game.AccAmountPeace.AccBet)
	}
	log.Infof("日志核对, 拔河第%d期, 获胜:%s,num:%f,numA:%d,numB:%d, 本期下注总额:%d, 本期赔付总额:%d, 历史反水比例:%f, 可赔付库存:%d, 庄家抽水:%d", game.RoundNum, hit, num, numA, numB, betTotalAmount, lossAmount, backWater, game.Amount, game.TakeAmount)
	return nil
}

func open(game *compete.Game, timestamp int64) error {
	log.Debug("拔河游戏开始")
	game.RoundNum++
	game.CurRound = &compete.Round{
		Number:     game.RoundNum,
		Orders:     map[string]*compete.Order{},
		StartTime:  timestamp,
		CloseTime:  timestamp + int64(utils.LubanTables.TBCompete.Get("open_time").NumInt*1000),
		ResultTime: timestamp + int64(utils.LubanTables.TBCompete.Get("open_time").NumInt+utils.LubanTables.TBCompete.Get("close_time").NumInt)*1000,
		EndTime:    timestamp + int64(utils.LubanTables.TBCompete.Get("open_time").NumInt+utils.LubanTables.TBCompete.Get("close_time").NumInt+utils.LubanTables.TBCompete.Get("settlement_time").NumInt)*1000,
	}
	game.State = 0
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_GameStageNotify_,
		Message: &pb.ClientResponse_GameStageNotify{GameStageNotify: &pb.GameStageNotify{
			Stage:     game.State,
			Countdown: game.Countdown(currentTime),
			Timestamp: currentTime,
		}},
	})
	return nil
}

func close(game *compete.Game) error {
	log.Debug("拔河游戏封盘")
	game.State = 1

	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_GameStageNotify_,
		Message: &pb.ClientResponse_GameStageNotify{
			GameStageNotify: &pb.GameStageNotify{
				Stage:     game.State,
				Countdown: game.Countdown(currentTime),
				Timestamp: currentTime,
			}},
	})
	return nil
}
