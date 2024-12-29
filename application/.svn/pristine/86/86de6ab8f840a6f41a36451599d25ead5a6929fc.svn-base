package squid

import (
	"application/api/handlers/schedule"
	"application/api/presenter"
	rechargedb "application/api/presenter/recharge"
	"application/api/presenter/squid"
	"application/mongodb"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"application/session"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

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

		game := &squid.Game{}
		if e := mongodb.Find(context.Background(), game, 0); e != nil {
			log.Error(e)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1005,
			})
		}

		globalSquidRound := &squid.GlobalRound{}
		err = mongodb.Find(context.Background(), globalSquidRound, userInfo.Squid.RoundId)
		if err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1006,
					Message: "Player did not join the game round",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1007,
			})
		}
		roundPlayers := globalSquidRound.Track1.PlayerNums + globalSquidRound.Track2.PlayerNums + globalSquidRound.Track3.PlayerNums + globalSquidRound.Track4.PlayerNums
		ts, _, err := redis.BPop(redis.SquidPrefix)
		if err != nil {
			log.Error(err)
		}
		Countdown := game.SquidCountdown(ts)
		if Countdown < 0 {
			Countdown = 0
		}

		squidFund := &squid.Fund{}
		if err := mongodb.Find(context.Background(), squidFund, 0); err != nil {
			log.Error(err)
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"curRoundId":   userInfo.Squid.RoundId,
				"game_stage":   game.State,
				"countdown":    Countdown,
				"balance":      userInfo.Balance,
				"roundPlayers": roundPlayers,
				"deadTrack":    game.DeadTrack,
				"jackpot":      squidFund.Jackpot,
			},
		})
	}
}

func Order() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type betRequest struct {
			SessionToken string `json:"sessionToken"`
			Track        int32  `json:"track"`  //选择赛道(1,2,3,4)
			Type         int32  `json:"type"`   //筹码类型(1,2,3,4,5)
			Amount       int32  `json:"amount"` //购买数量
		}
		var bet betRequest
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
		account, err := redis.GetAccountBySessionToken(bet.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "sessionToken error",
			})
		}
		if bet.Track < 1 || bet.Track > 4 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "param error",
			})
		}
		game := &squid.Game{}
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
		if userInfo.Squid.RoundId > squid.TotalRounds || userInfo.Squid.RoundId < 0 {
			log.Error("RoundId error: ", userInfo.Account, userInfo.Squid.RoundId)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1008,
			})
		}
		if userInfo.Squid.Track != 0 && userInfo.Squid.Track != bet.Track { //可加注
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1009,
				Message: "Selected tracks for this round",
			})
		}
		if bet.Amount == 0 {
			bet.Amount = 1
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
				Code:    1010,
				Message: "param error",
			})
		}
		price := int64(utils.LubanTables.TBWood.Get(priceLabel).NumInt) * int64(bet.Amount)

		//检查余额
		if !mongodb.CheckBalance(userInfo, price) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1011,
				Message: "balance not enough",
			})
		}

		if userInfo.Squid.RoundId == 0 { //玩家Round值在结算时更新, 首次进入游戏、死亡、退赛置1
			userInfo.Squid.RoundId = 1
		}

		if err := mongodb.SquidOrder(userInfo, price, bet.Track); err != nil {
			log.Error("SquidOrder error: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1012,
			})
		}

		// 添加下注的玩家账号 ID 到 Redis
		if err := redis.AddSquidPlayer(account); err != nil {
			log.Error("Failed to add player ID to Redis:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1013,
			})
		}

		// 通知client
		if err := sendSquidBetInfoNotify(userInfo.Squid.RoundId, pb.SquidBetInfoEnumType_OrderType); err != nil {
			log.Error("sendSquidBetInfoNotify err: ", err)
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"balance":           userInfo.Balance,
				"roundId":           userInfo.Squid.RoundId,
				"track":             userInfo.Squid.Track,
				"bet_prices":        userInfo.Squid.BetPrices,
				"BetPricesPerRound": userInfo.Squid.BetPricesPerRound,
			},
		})
	}
}

func Cancel() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type cancelRequest struct {
			SessionToken string `json:"sessionToken"`
		}
		var cancel cancelRequest
		if err := c.BodyParser(&cancel); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if cancel.SessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
			})
		}
		account, err := redis.GetAccountBySessionToken(cancel.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "sessionToken error",
			})
		}
		game := &squid.Game{}
		if e := mongodb.Find(context.Background(), game, 0); e != nil {
			log.Error(e)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1003,
			})
		}
		if game.State != 0 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1004,
				Message: "Closed",
			})
		}
		userInfo := &presenter.UserInfo{}
		err = mongodb.Find(context.Background(), userInfo, account)
		if err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1005,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1006,
			})
		}
		curRoundId := userInfo.Squid.RoundId
		if curRoundId == 0 || curRoundId > squid.TotalRounds {
			log.Error("RoundId error: ", userInfo.Account, userInfo.Squid.RoundId)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1007,
			})
		}

		// db更新account: 重置本轮下注额, 返还余额;
		// db更新squid_round
		if err := mongodb.SquidCancel(userInfo); err != nil {
			log.Error("SquidCancel error: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1008,
			})
		}

		// 从 Redis 删除玩家账号 ID
		if err := redis.RemoveSquidPlayer(account); err != nil {
			log.Error("Failed to remove player ID from Redis:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1009,
			})
		}

		// 通知client
		if err := sendSquidBetInfoNotify(curRoundId, pb.SquidBetInfoEnumType_CancelType); err != nil {
			log.Error("sendSquidBetInfoNotify err: ", err)
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"balance":           userInfo.Balance,
				"roundId":           userInfo.Squid.RoundId,
				"track":             userInfo.Squid.Track,
				"bet_prices":        userInfo.Squid.BetPrices,
				"BetPricesPerRound": userInfo.Squid.BetPricesPerRound,
			},
		})
	}
}

func Switch() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type switchRequest struct {
			SessionToken string `json:"sessionToken"`
			Track        int32  `json:"track"` //要换的赛道(1,2,3,4)
		}
		var switchTrack switchRequest
		if err := c.BodyParser(&switchTrack); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if switchTrack.SessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
			})
		}
		account, err := redis.GetAccountBySessionToken(switchTrack.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "sessionToken error",
			})
		}
		if switchTrack.Track < 1 || switchTrack.Track > 4 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "Track parameter error",
			})
		}
		game := &squid.Game{}
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
		if userInfo.Squid.RoundId > squid.TotalRounds || userInfo.Squid.RoundId < 0 {
			log.Error("RoundId error: ", userInfo.Account, userInfo.Squid.RoundId)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1008,
			})
		}
		if userInfo.Squid.Track == 0 && userInfo.Squid.BetPrices == 0 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1009,
				Message: "Please place your bet first",
			})
		}
		if userInfo.Squid.Track == switchTrack.Track {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1010,
				Message: "Already on track",
			})
		}

		// db更新account: 转换赛道
		if err := mongodb.SquidSwitch(userInfo, switchTrack.Track); err != nil {
			log.Error("SquidSwitch error: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1011,
				Message: "Failed to SquidSwitch",
			})
		}

		// 通知client
		if err := sendSquidBetInfoNotify(userInfo.Squid.RoundId, pb.SquidBetInfoEnumType_SwitchType); err != nil {
			log.Error("sendSquidBetInfoNotify err: ", err)
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"balance":           userInfo.Balance,
				"roundId":           userInfo.Squid.RoundId,
				"track":             userInfo.Squid.Track,
				"bet_prices":        userInfo.Squid.BetPrices,
				"BetPricesPerRound": userInfo.Squid.BetPricesPerRound,
			},
		})
	}
}
func sendSquidBetInfoNotify(curRoundId int32, notifyType pb.SquidBetInfoEnumType) error {
	globalSquidRound := &squid.GlobalRound{}
	err := mongodb.Find(context.Background(), globalSquidRound, curRoundId)
	if err != nil {
		log.Error(err)
		return err
	}
	trackRefs := []*squid.Track{
		&globalSquidRound.Track1,
		&globalSquidRound.Track2,
		&globalSquidRound.Track3,
		&globalSquidRound.Track4,
	}
	tracks := make([]*pb.Track, squid.TotalTracks)
	for i, trackRef := range trackRefs {
		tracks[i] = &pb.Track{
			TrackId: int32(i + 1),
			Players: trackRef.PlayerNums,
			Bets:    trackRef.TotalBetPrices + trackRef.RobotTotalBetPrices,
		}
	}

	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_SquidBetInfoNotify_,
		Message: &pb.ClientResponse_SquidBetInfoNotify{SquidBetInfoNotify: &pb.SquidBetInfoNotify{
			RoundId: curRoundId,
			Type:    notifyType,
			Track:   tracks,
		}},
	})
	return nil
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
		historyOrders, err := mongodb.FindSquidUserOrders(context.Background(), account)
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

func Data() fiber.Handler {
	return func(c *fiber.Ctx) error {
		latest, results, err := redis.GetSquidData()
		if err != nil {
			log.Error("GetSquidAllData error: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1000,
			})
		}
		schedule.SendToDingTalk()
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"latest":  latest,
				"results": results,
			},
		})
	}
}

func FirstPassStatus() fiber.Handler {
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
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"lastFirstPassDate": userInfo.Squid.FirstPass.LastFirstPassDate,
				"pool":              userInfo.Squid.FirstPass.Pool,
			},
		})
	}
}

func UsdtToSqu() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
			Amount       int64  `json:"amount"` // 美分数量
		}
		var req request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if req.SessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
			})
		}
		if req.Amount < 100 { //至少1美元
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "params error",
			})
		}
		account, err := redis.GetAccountBySessionToken(req.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "sessionToken error",
			})
		}
		userInfo := &presenter.UserInfo{}
		if err = mongodb.Find(context.Background(), userInfo, account); err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1004,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1005,
			})
		}
		if !mongodb.CheckUSDT(userInfo, req.Amount) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1006,
				Message: "USDT not enough",
			})
		}

		// db更新account: USDT 转换为 游戏币和兑换券
		if err := mongodb.UsdtToSqu(userInfo, req.Amount); err != nil {
			log.Error("UsdtToSqu error: ", err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code: 1007,
			})
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"balance": userInfo.Balance,
				"usdt":    userInfo.USDT,
				"voucher": userInfo.Voucher,
			},
		})
	}
}

func SquToUsdt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
			Squ          int64  `json:"squ"` // 游戏币数量 (分为单位)
			Vou          int64  `json:"vou"` // 兑换券数量
		}
		var req request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if req.SessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
			})
		}
		if req.Squ < 1 || req.Vou < 1 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "params error",
			})
		}
		squPerUsdt := int64(utils.LubanTables.TBApp.Get("allocation").NumInt)  // USDT 转换 游戏币汇率 (1美元:1000000)
		vouPerUsdt := int64(utils.LubanTables.TBApp.Get("allocatioon").NumInt) // USDT 转换 兑换券汇率 (1美元:10)
		// 校验 Squ 和 Vou 的比例是否正确
		expectedSqu := req.Vou * (squPerUsdt / vouPerUsdt)
		if req.Squ != expectedSqu {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "Invalid Squ to Vou ratio",
			})
		}
		account, err := redis.GetAccountBySessionToken(req.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1004,
				Message: "sessionToken error",
			})
		}
		userInfo := &presenter.UserInfo{}
		if err = mongodb.Find(context.Background(), userInfo, account); err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1005,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1006,
			})
		}
		if !mongodb.CheckBalance(userInfo, req.Squ) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1007,
				Message: "Squ not enough",
			})
		}
		if !mongodb.CheckVoucher(userInfo, req.Vou) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1008,
				Message: "Vou not enough",
			})
		}
		oldBalance := userInfo.Balance
		oldUSDT := userInfo.USDT
		oldVoucher := userInfo.Voucher

		// 游戏币和兑换券 转换为 USDT
		mongodb.DecrAmount(userInfo, req.Squ)
		mongodb.DecrVoucher(userInfo, req.Vou)
		// 转为美分
		usdtToAdd := req.Vou * 100 / vouPerUsdt // req.Squ/squPerUsdt*100
		mongodb.AddUSDT(userInfo, usdtToAdd)
		if err = mongodb.Update(context.Background(), userInfo, nil); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1009,
			})
		}

		exchangeInfo := &rechargedb.ExchangeInfo{
			Account:   userInfo.Account,
			Type:      rechargedb.SquToUsdt,
			Amount:    usdtToAdd,
			Coin:      req.Squ,
			Voucher:   req.Vou,
			CreatedAt: time.Now().UnixMilli(),
		}
		if err := mongodb.Insert(context.Background(), exchangeInfo); err != nil {
			log.Error(err)
			return err
		}

		fund := &rechargedb.Fund{}
		if err := mongodb.Find(context.Background(), fund, 0); err != nil {
			log.Error(err)
			return err
		}
		fund.SquToUsdtCoins += req.Squ
		if e := mongodb.Update(context.Background(), fund, nil); e != nil {
			log.Error("Error updating recharge fund:", e)
		}

		log.Infof("用户:%s, 游戏币/兑换券==>美分, 旧美分:%d,兑换了:%d,新美分:%d, 旧游戏币:%d,消耗了:%d,新游戏币:%d, 旧券%d,消耗了:%d,新券:%d",
			userInfo.Account, oldUSDT, usdtToAdd, userInfo.USDT, oldBalance, req.Squ, userInfo.Balance, oldVoucher, req.Vou, userInfo.Voucher)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"balance": userInfo.Balance,
				"usdt":    userInfo.USDT,
				"voucher": userInfo.Voucher,
			},
		})
	}
}
