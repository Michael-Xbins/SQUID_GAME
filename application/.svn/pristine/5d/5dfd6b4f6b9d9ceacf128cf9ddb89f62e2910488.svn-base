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
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"math"
	"strconv"
	"strings"
	"time"
)

const scaleFactor = int64(1000) // 千分比,保持赔率的精度

func Order() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var bet presenter.LadderBet
		if err := c.BodyParser(&bet); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if bet.SessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
			})
		}
		if len(bet.Orders) > 22 || len(bet.Orders) <= 0 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "param error",
			})
		}
		for _, order := range bet.Orders {
			if len(order.Infos) == 0 {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1003,
					Message: "param error",
				})
			}
			for _, info := range order.Infos {
				if info.Type < 1 || info.Type > 5 {
					return c.Status(fiber.StatusOK).JSON(presenter.Response{
						Code:    1004,
						Message: "param error",
					})
				}
				if info.Num <= 0 {
					return c.Status(fiber.StatusBadRequest).JSON(presenter.Response{
						Code:    1005,
						Message: "param error",
					})
				}
			}
		}

		account, err := redis.GetAccountBySessionToken(bet.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1006,
				Message: "sessionToken error",
			})
		}
		game := &ladder.Game{}
		if e := mongodb.Find(context.Background(), game, 0); e != nil {
			log.Error(e)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1007,
			})
		}
		if game.State != 0 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1008,
				Message: "Closed",
			})
		}
		userInfo := &presenter.UserInfo{}
		err = mongodb.Find(context.Background(), userInfo, account)
		if err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1009,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1010,
			})
		}

		limitMap := map[string]int64{
			"lim_1":  int64(utils.LubanTables.TBLadder.Get("lim_1").NumInt),
			"lim_2":  int64(utils.LubanTables.TBLadder.Get("lim_2").NumInt),
			"lim_3":  int64(utils.LubanTables.TBLadder.Get("lim_3").NumInt),
			"lim_23": int64(utils.LubanTables.TBLadder.Get("lim_23").NumInt),
		}
		priceMap := map[int32]string{
			1: "price_one",
			2: "price_two",
			3: "price_three",
			4: "price_four",
			5: "price_five",
		}
		curBets, err := redis.GetLadderBets(c.Context(), game.RoundNum)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1011,
				Message: "Failed to retrieve bets",
			})
		}
		totalPrice := int64(0) //所有订单总额
		accBet1 := int64(0)
		accBet2 := int64(0)
		var orders []*ladder.Order
		betIdPrices := make(map[string]int64) // 用于存储每个 BetId 的总下注额
		for _, order := range bet.Orders {
			ladderItem := utils.LubanTables.TBLadder.Get(order.Id)
			if ladderItem == nil {
				return c.Status(fiber.StatusBadRequest).JSON(presenter.Response{
					Code:    1012,
					Message: "Invalid order ID",
				})
			}
			limitGroup := getLimitGroup(order.Id)
			if limitGroup == "" {
				return c.Status(fiber.StatusBadRequest).JSON(presenter.Response{
					Code:    1013,
					Message: "Invalid order ID",
				})
			}
			category := ladderItem.Category
			orderPrice := int64(0)

			for _, info := range order.Infos {
				priceLabel, ok := priceMap[info.Type]
				if !ok {
					return c.Status(fiber.StatusOK).JSON(presenter.Response{
						Code:    1014,
						Message: "param error",
					})
				}
				price := int64(utils.LubanTables.TBLadder.Get(priceLabel).NumInt * info.Num)
				totalPrice += price
				orderPrice += price
				if category == 1 {
					accBet1 += price
				} else if category == 2 {
					accBet2 += price
				}
			}

			// 检查是否超过限红
			if curBets[order.Id]+orderPrice > limitMap[limitGroup] {
				return c.Status(fiber.StatusBadRequest).JSON(presenter.Response{
					Code:    1015,
					Message: "Exceeding the maximum limit of the sub-disk",
				})
			}

			odd := int64(utils.LubanTables.TBLadder.Get(order.Id).NumInt)
			realOdd := float64(odd) / float64(scaleFactor)
			orderInfo := &ladder.Order{
				OrderId:    uuid.New().String(),
				Account:    account,
				RoundNum:   game.RoundNum,
				BetId:      order.Id,
				Category:   category,
				Infos:      order.Infos,
				OrderPrice: orderPrice,
				Odds:       math.Floor(realOdd*100) / 100,
				Bonus:      int64(float64(orderPrice) * math.Floor(realOdd*100) / 100),
				Timestamp:  time.Now().UnixMilli(),
			}
			orders = append(orders, orderInfo)
			betIdPrices[order.Id] += orderPrice
		}

		// 检查余额
		if !mongodb.CheckBalance(userInfo, totalPrice) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1016,
				Message: "balance not enough",
			})
		}
		// 更新余额
		mongodb.DecrAmount(userInfo, totalPrice)
		// 更新玩家的流水额(下注的钱)
		mongodb.AddLadderTurnOver(userInfo, totalPrice)
		// 存储订单
		if err := mongodb.InsertMany(context.Background(), orders); err != nil {
			log.Error("insert orders error: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1017,
			})
		}
		pumpDetails := getPumpDetails(totalPrice)

		// Redis 分布式锁, 更新全局ladderFund
		lockKey := redis.LadderPrefix + "fund_lock"
		if err := redis.TryLock(c.Context(), lockKey, 10); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1018,
				Message: "Failed to acquire lock, please try again later",
			})
		}
		defer func() {
			if err := redis.UnLock(c.Context(), lockKey); err != nil {
				log.Error("Failed to release lock: ", err)
			}
		}()
		// 更新抽水
		ladderFund := &ladder.Fund{}
		if e := mongodb.Find(context.Background(), ladderFund, 0); e != nil {
			log.Error(e)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1019,
			})
		}
		mongodb.AddLadderHouseCut(ladderFund, pumpDetails.HouseCut)
		mongodb.AddLadderAvailableFund(ladderFund, pumpDetails.AvailableContribution)
		mongodb.AddLadderAgent(userInfo, ladderFund, pumpDetails)
		ladderFund.AccAmount1.AccBet += accBet1
		ladderFund.AccAmount2.AccBet += accBet2

		// 添加下注的玩家账号 ID 到 Redis
		if err := redis.AddLadderPlayer(account, game.RoundNum); err != nil {
			log.Error("Failed to add player ID to Redis:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1020,
			})
		}

		for orderID, amount := range betIdPrices {
			if err := redis.IncrementLadderBets(c.Context(), game.RoundNum, orderID, amount); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code:    1021,
					Message: "Failed to update bet amount",
				})
			}
		}

		err = mongodb.Update(context.Background(), userInfo, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1022,
			})
		}
		err = mongodb.Update(context.Background(), ladderFund, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1023,
			})
		}

		log.Debugf("梯子下注, account:%s 本次所有订单: 第%d期, 总注额:%d, upline: %v, oldHouseCut: %v,  本次抽水 庄家: %v, 可赔付库存: %v, 代理: %v,代理上线:%v,代理上上线:%v",
			userInfo.Account, game.RoundNum, totalPrice, userInfo.UpLine, ladderFund.HouseCut, pumpDetails.HouseCut, pumpDetails.AvailableContribution,
			pumpDetails.AgentContribution, pumpDetails.UpLineContribution, pumpDetails.UpUpLineContribution)

		// 通知所有client
		newBets, err := redis.GetLadderBets(c.Context(), game.RoundNum)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1024,
				Message: "Failed to retrieve bets",
			})
		}
		session.S2AllMessage(&pb.ClientResponse{
			Type: pb.MessageType_LadderBetInfoNotify_,
			Message: &pb.ClientResponse_LadderBetInfoNotify{LadderBetInfoNotify: &pb.LadderBetInfoNotify{
				Bets: newBets,
			}},
		})

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"balance":  userInfo.Balance,
				"roundNum": game.RoundNum,
				"bets":     newBets,
			},
		})
	}
}

// 根据轮次ID确定限红组
func getLimitGroup(roundID string) string {
	if strings.HasPrefix(roundID, "round_") {
		numStr := strings.TrimPrefix(roundID, "round_")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			log.Errorf("Failed to parse round number from roundID %s: %v", roundID, err)
			return ""
		}

		switch {
		case num >= 10 && num <= 12:
			return "lim_1"
		case num >= 20 && num <= 23:
			return "lim_2"
		case num >= 30 && num <= 32:
			return "lim_3"
		case num >= 2030 && num <= 2332:
			return "lim_23"
		}
	}
	return ""
}

func getPumpDetails(userRoundBet int64) *presenter.PumpDetails { // 千分比,保持赔率的精度
	const scaleFactor = int64(1000)                                           // 千分比,保持赔率的精度
	pumpProfit := int64(utils.LubanTables.TBLadder.Get("pump_profit").NumInt) //千分比
	pumpActing := int64(utils.LubanTables.TBLadder.Get("pump_acting").NumInt) //千分比
	acting0 := int64(utils.LubanTables.TBLadder.Get("acting0").NumInt)        //千分比
	acting1 := int64(utils.LubanTables.TBLadder.Get("acting1").NumInt)        //千分比

	agent := userRoundBet * pumpActing / scaleFactor
	upLineAgent := userRoundBet * pumpActing * acting0 / scaleFactor / scaleFactor
	upUpLineAgent := agent - upLineAgent

	details := &presenter.PumpDetails{
		PumpProfit:            pumpProfit,
		PumpActing:            pumpActing,
		Acting0:               acting0,
		Acting1:               acting1,
		HouseCut:              userRoundBet * pumpProfit / scaleFactor,
		AgentContribution:     agent,
		UpLineContribution:    upLineAgent,
		UpUpLineContribution:  upUpLineAgent,
		AvailableContribution: userRoundBet * (scaleFactor - pumpProfit - pumpActing) / scaleFactor,
	}
	return details
}

func HistoryOrders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type statusRequest struct {
			SessionToken string `json:"sessionToken"`
		}
		var status statusRequest
		if err := c.BodyParser(&status); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if status.SessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
			})
		}
		account, err := redis.GetAccountBySessionToken(status.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "sessionToken error",
			})
		}
		historyOrders, err := mongodb.FindLadderUserOrders(context.Background(), account, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1003,
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"historyOrders": historyOrders,
			},
		})
	}
}

func Lottery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		lotteries, err := mongodb.FindLadderLotteries(context.Background())
		if err != nil {
			log.Error("FindAllLotteries error: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1000,
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"lotteries": lotteries,
			},
		})
	}
}

func Status() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type statusRequest struct {
			SessionToken string `json:"sessionToken"`
		}
		var status statusRequest
		if err := c.BodyParser(&status); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if status.SessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
			})
		}
		account, err := redis.GetAccountBySessionToken(status.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "sessionToken error",
			})
		}
		userInfo := &presenter.UserInfo{}
		err = mongodb.Find(context.Background(), userInfo, account)
		if err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1003,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1004,
			})
		}

		game := &ladder.Game{}
		if e := mongodb.Find(context.Background(), game, 0); e != nil {
			log.Error(e)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1005,
			})
		}
		timestamp := time.Now().UnixMilli()
		Countdown := game.LadderCountdown(timestamp)
		bets, err := redis.GetLadderBets(c.Context(), game.RoundNum)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1016,
				Message: "Failed to retrieve bets",
			})
		}
		lottery := &ladder.Lottery{}
		_ = mongodb.Find(context.Background(), lottery, game.RoundNum)

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"roundNum":   game.RoundNum,
				"game_stage": game.State,
				"countdown":  Countdown,
				"balance":    userInfo.Balance,
				"transHash":  lottery.Hash,
				"bets":       bets,
			},
		})
	}
}

type AccAmount struct {
	AccBet    int64   `json:"accBet"`    // 游戏累计投注额
	AccPayout int64   `json:"accPayout"` // 游戏累计赔付额
	Rate      float64 `json:"rate"`      // 反水比例
}

func GetData() (*AccAmount, *AccAmount, int64, error) {
	ladderFund := &ladder.Fund{}
	if e := mongodb.Find(context.Background(), ladderFund, 0); e != nil {
		log.Error(e)
		return nil, nil, 0, e
	}
	accAmount1 := &AccAmount{
		AccBet:    ladderFund.AccAmount1.AccBet,
		AccPayout: ladderFund.AccAmount1.AccPayout,
	}
	accAmount2 := &AccAmount{
		AccBet:    ladderFund.AccAmount2.AccBet,
		AccPayout: ladderFund.AccAmount2.AccPayout,
	}

	if accAmount1.AccBet > 0 {
		accAmount1.Rate = float64(accAmount1.AccPayout) / float64(accAmount1.AccBet)
	}
	if accAmount2.AccBet > 0 {
		accAmount2.Rate = float64(accAmount2.AccPayout) / float64(accAmount2.AccBet)
	}

	return accAmount1, accAmount2, ladderFund.AvailableFund, nil
}
func Data() fiber.Handler {
	return func(c *fiber.Ctx) error {
		accAmount1, accAmount2, AvailableFund, err := GetData()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1000,
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"TrackType1":    accAmount1,
				"TrackType2":    accAmount2,
				"AvailableFund": AvailableFund,
			},
		})
	}
}
