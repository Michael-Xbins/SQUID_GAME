package glass

import (
	"application/api/presenter"
	"application/api/presenter/glass"
	"application/mongodb"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"math"
	"strconv"
	"time"
)

func Order() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var bet presenter.GlassBet
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
		if len(bet.Tracks) > 3 || len(bet.Tracks) <= 0 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "Track parameter error",
			})
		}
		account, err := redis.GetAccountBySessionToken(bet.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "sessionToken error",
			})
		}
		game := &glass.Game{}
		if e := mongodb.Find(context.Background(), game, 0); e != nil {
			log.Error(e)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1004,
			})
		}
		if game.State != 0 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1005,
				Message: "Closed",
			})
		}
		userInfo := &presenter.UserInfo{}
		err = mongodb.Find(context.Background(), userInfo, account)
		if err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1006,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1007,
			})
		}
		priceMap := map[int32]string{
			1: "price_one",
			2: "price_two",
			3: "price_three",
			4: "price_four",
			5: "price_five",
		}
		priceLabel, ok := priceMap[bet.Type]
		if !ok {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1008,
				Message: "Chip parameter error",
			})
		}
		price := int64(utils.LubanTables.TBGlass.Get(priceLabel).NumInt)
		//检查余额
		if !mongodb.CheckBalance(userInfo, price) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1009,
				Message: "balance not enough",
			})
		}

		trackIds := make(map[int32]bool)
		allowedNums := map[int32][]int{
			1: {0, 1, 2},
			2: {0, 1, 2, 3},
			3: {0, 1, 2},
		}
		odds := int64(1)
		scaleFactor := int64(1)
		var track1, track2, track3 string
		for _, track := range bet.Tracks {
			if _, exists := trackIds[track.Id]; exists {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1010,
					Message: "Duplicate track ID",
				})
			}
			if !contains(allowedNums[track.Id], track.Num) {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1011,
					Message: "Not within the permitted range",
				})
			}
			trackIds[track.Id] = true

			switch track.Id {
			case 1:
				track1 = strconv.Itoa(track.Num)
			case 2:
				track2 = strconv.Itoa(track.Num)
			case 3:
				track3 = strconv.Itoa(track.Num)
			}
			key := fmt.Sprintf("round_%d%d", track.Id, track.Num)
			odd := int64(utils.LubanTables.TBGlass.Get(key).NumInt)
			odds *= odd
			scaleFactor *= 1000
		}

		// 更新余额
		mongodb.DecrAmount(userInfo, price)
		// 更新玩家的流水额(下注的钱)
		mongodb.AddGlassTurnOver(userInfo, price)

		// Redis 分布式锁, 更新全局glassFund
		lockKey := redis.GlassPrefix + "fund_lock"
		if err := redis.TryLock(c.Context(), lockKey, 10); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1012,
				Message: "Failed to acquire lock, please try again later",
			})
		}
		defer func() {
			if err := redis.UnLock(c.Context(), lockKey); err != nil {
				log.Error("Failed to release lock: ", err)
			}
		}()
		// 更新抽水
		glassFund := &glass.Fund{}
		if e := mongodb.Find(context.Background(), glassFund, 0); e != nil {
			log.Error(e)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1013,
			})
		}
		pumpDetails := getPumpDetails(price)
		glassFund.HouseCut += pumpDetails.HouseCut
		glassFund.AvailableFund += pumpDetails.AvailableContribution
		mongodb.AddGlassAgent(userInfo, glassFund, pumpDetails)
		if len(bet.Tracks) == 1 {
			glassFund.AccAmount1.AccBet += price
		} else if len(bet.Tracks) == 2 {
			glassFund.AccAmount2.AccBet += price
		} else {
			glassFund.AccAmount3.AccBet += price
		}

		// 订单
		realOdds := float64(odds) / float64(scaleFactor)
		orderInfo := &glass.Order{
			OrderId:   uuid.New().String(),
			Account:   account,
			RoundNum:  game.RoundNum,
			BetType:   priceLabel,
			TrackType: len(bet.Tracks),
			Track1:    track1,
			Track2:    track2,
			Track3:    track3,
			Odds:      math.Floor(realOdds*100) / 100, // 保留两位小数并向下取整
			Timestamp: time.Now().UnixMilli(),
		}
		if err := mongodb.Insert(context.Background(), orderInfo); err != nil {
			log.Error("insert order error: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1014,
			})
		}

		err = mongodb.Update(context.Background(), userInfo, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1015,
			})
		}
		err = mongodb.Update(context.Background(), glassFund, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1016,
			})
		}

		// 添加下注的玩家账号 ID 到 Redis
		if err := redis.AddGlassPlayer(account); err != nil {
			log.Error("Failed to add player ID to Redis:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1017,
			})
		}

		log.Debugf("玻璃桥下注：account: %v, odds: %v, upline: %v, oldHouseCut: %v,  本次抽水 庄家: %v, 可赔付库存: %v, 代理: %v,代理上线:%v,代理上上线:%v",
			userInfo.Account, realOdds, userInfo.UpLine, glassFund.HouseCut, pumpDetails.HouseCut, pumpDetails.AvailableContribution,
			pumpDetails.AgentContribution, pumpDetails.UpLineContribution, pumpDetails.UpUpLineContribution)

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"balance":  userInfo.Balance,
				"roundNum": game.RoundNum,
				"odds":     math.Floor(realOdds*100) / 100,
			},
		})
	}
}
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
func getPumpDetails(userRoundBet int64) *presenter.PumpDetails { // 千分比,保持赔率的精度
	const scaleFactor = int64(1000)                                          // 千分比,保持赔率的精度
	pumpProfit := int64(utils.LubanTables.TBGlass.Get("pump_profit").NumInt) //千分比
	pumpActing := int64(utils.LubanTables.TBGlass.Get("pump_acting").NumInt) //千分比
	acting0 := int64(utils.LubanTables.TBGlass.Get("acting0").NumInt)        //千分比
	acting1 := int64(utils.LubanTables.TBGlass.Get("acting1").NumInt)        //千分比

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
		historyOrders, err := mongodb.FindGlassUserOrders(context.Background(), account, nil)
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
		lotteries, err := mongodb.FindAllLotteries(context.Background())
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

		game := &glass.Game{}
		if e := mongodb.Find(context.Background(), game, 0); e != nil {
			log.Error(e)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1005,
			})
		}
		timestamp := time.Now().UnixMilli()
		Countdown := game.GlassCountdown(timestamp)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"roundNum":   game.RoundNum,
				"game_stage": game.State,
				"countdown":  Countdown,
				"balance":    userInfo.Balance,
				"transHash":  game.TransHash,
			},
		})
	}
}

type AccAmount struct {
	AccBet    int64   `json:"accBet"`    // 游戏累计投注额
	AccPayout int64   `json:"accPayout"` // 游戏累计赔付额
	Rate      float64 `json:"rate"`      // 反水比例
}

func GetData() (*AccAmount, *AccAmount, *AccAmount, int64, error) {
	glassFund := &glass.Fund{}
	if e := mongodb.Find(context.Background(), glassFund, 0); e != nil {
		log.Error(e)
		return nil, nil, nil, 0, e
	}
	accAmount1 := &AccAmount{
		AccBet:    glassFund.AccAmount1.AccBet,
		AccPayout: glassFund.AccAmount1.AccPayout,
	}
	accAmount2 := &AccAmount{
		AccBet:    glassFund.AccAmount2.AccBet,
		AccPayout: glassFund.AccAmount2.AccPayout,
	}
	accAmount3 := &AccAmount{
		AccBet:    glassFund.AccAmount3.AccBet,
		AccPayout: glassFund.AccAmount3.AccPayout,
	}
	if accAmount1.AccBet > 0 {
		accAmount1.Rate = float64(accAmount1.AccPayout) / float64(accAmount1.AccBet)
	}
	if accAmount2.AccBet > 0 {
		accAmount2.Rate = float64(accAmount2.AccPayout) / float64(accAmount2.AccBet)
	}
	if accAmount3.AccBet > 0 {
		accAmount3.Rate = float64(accAmount3.AccPayout) / float64(accAmount3.AccBet)
	}
	return accAmount1, accAmount2, accAmount3, glassFund.AvailableFund, nil
}
func Data() fiber.Handler {
	return func(c *fiber.Ctx) error {
		accAmount1, accAmount2, accAmount3, AvailableFund, err := GetData()
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
				"TrackType3":    accAmount3,
				"AvailableFund": AvailableFund,
			},
		})
	}
}
