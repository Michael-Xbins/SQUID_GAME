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
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func Order(c *fiber.Ctx) error {
	type OrderRequest struct {
		SessionToken string `json:"sessionToken"`
		PriceType    string `json:"priceType"`
		Param        string `json:"param"`
	}
	var request OrderRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1000,
			Message: "Invalid request body",
		})
	}
	priceType := request.PriceType
	param := request.Param
	if request.SessionToken == "" {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1001,
			Message: "sessionToken is required",
		})
	}
	account, err := redis.GetAccountBySessionToken(request.SessionToken)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1002,
			Message: "sessionToken error",
		})
	}
	if priceType != "price_one" && priceType != "price_two" && priceType != "price_three" && priceType != "price_four" && priceType != "price_five" {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1003,
			Message: "priceType error",
		})
	}
	price := int64(utils.LubanTables.TBCompete.Get(priceType).NumInt)
	if price == 0 {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1004,
			Message: "price error",
		})
	}
	if param != "a" && param != "b" && param != "peace" {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1005,
			Message: "param error",
		})
	}

	userInfo := &presenter.UserInfo{}
	err = mongodb.Find(context.Background(), userInfo, account)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1007,
		})
	}
	game := &compete.Game{}
	if e := mongodb.Find(context.Background(), game, 0); e != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1008,
		})
	}
	if game.State != 0 {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1009,
			Message: "game closed",
		})
	}
	//if game.Amount <= 0 {
	//	return c.Status(fiber.StatusOK).JSON(presenter.Response{
	//		Code:    1010,
	//		Message: "game over",
	//	})
	//}
	if !mongodb.CheckBalance(userInfo, price) {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1011,
			Message: "balance not enough",
		})
	}
	mongodb.DecrAmount(userInfo, price)

	userInfo.CompeteLastBet += price //拔河本轮下注额(暂存), 用于核对账单, 结算/取消清零

	round := game.CurRound

	order, ok := round.Orders[account]
	if !ok {
		order = &compete.Order{
			Amounts: map[string]int64{},
		}
		round.Orders[account] = order
	}
	order.Amounts[param] += price

	var val int64
	totalAmounts := map[string]int64{}
	for _, v := range round.Orders {
		for tp, n := range v.Amounts {
			totalAmounts[tp] += n
		}
	}
	val = totalAmounts[param]
	if val > int64(utils.LubanTables.TBCompete.Get(fmt.Sprintf("lim_%s", param)).NumInt) {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1011,
			Message: "Exceeding the maximum limit of the sub-disk",
		})
	}

	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_TotalAmountChangeNotify_,
		Message: &pb.ClientResponse_TotalAmountChangeNotify{TotalAmountChangeNotify: &pb.TotalAmountChangeNotify{
			TotalAmounts: totalAmounts,
			Type:         param,
			Account:      account,
		}},
	})

	if err = mongodb.Update(context.Background(), userInfo, nil); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1013,
		})
	}
	if err = mongodb.Update(context.Background(), game, nil); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1014,
		})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.Response{
		Code: 1,
		Result: map[string]interface{}{
			"newBalance":   userInfo.Balance,
			"myAmounts":    order.Amounts,
			"totalAmounts": totalAmounts,
		},
	})
}

func Cancel(c *fiber.Ctx) error {
	type CancelRequest struct {
		SessionToken string `json:"sessionToken"`
	}
	var request CancelRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1000,
			Message: "Invalid request body",
		})
	}
	sessionToken := request.SessionToken
	if sessionToken == "" {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1001,
			Message: "Invalid session",
		})
	}
	account, err := redis.GetAccountBySessionToken(sessionToken)
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
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1003,
		})
	}
	game := &compete.Game{}
	if e := mongodb.Find(context.Background(), game, 0); e != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1004,
		})
	}
	if game.State != 0 {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1005,
			Message: "game closed",
		})
	}
	round := game.CurRound
	order, ok := round.Orders[account]
	if ok {
		totalAmount := int64(0)
		for _, v := range order.Amounts {
			totalAmount += v
		}
		if totalAmount > 0 {
			mongodb.AddAmount(userInfo, totalAmount)
			delete(round.Orders, account)
			if err = mongodb.Update(context.Background(), game, nil); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1006,
				})
			}
			userInfo.CompeteLastBet -= totalAmount //拔河本轮下注额(暂存), 用于核对账单, 结算/取消清零

			if err = mongodb.Update(context.Background(), userInfo, nil); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1007,
				})
			}
		}
	}
	totalAmounts := map[string]int64{}
	for _, v := range round.Orders {
		for tp, val := range v.Amounts {
			totalAmounts[tp] += val
		}
	}
	session.S2AllMessage(&pb.ClientResponse{
		Type:    pb.MessageType_TotalAmountChangeNotify_,
		Message: &pb.ClientResponse_TotalAmountChangeNotify{TotalAmountChangeNotify: &pb.TotalAmountChangeNotify{TotalAmounts: totalAmounts}},
	})
	return c.Status(fiber.StatusOK).JSON(presenter.Response{
		Code: 1,
		Result: map[string]interface{}{
			"totalAmounts": totalAmounts,
			"balance":      userInfo.Balance,
		},
	})
}

func History(c *fiber.Ctx) error {
	game := &compete.Game{}
	if e := mongodb.Find(context.Background(), game, 0); e != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1004,
		})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.Response{
		Code: 1,
		Result: map[string]interface{}{
			"resultHistory": game.ResultHistoryList,
		},
	})
}

func HistoryOrders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
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
		account, err := redis.GetAccountBySessionToken(req.SessionToken)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "sessionToken error",
			})
		}
		historyOrders, err := mongodb.FindCompeteUserOrders(context.Background(), account)
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

type AccAmount struct {
	AccBet    int64   `json:"accBet"`    // 游戏累计投注额
	AccPayout int64   `json:"accPayout"` // 游戏累计赔付额
	Rate      float64 `json:"rate"`      // 反水比例
}

func GetData() (*AccAmount, *AccAmount, *AccAmount, int64, error) {
	game := &compete.Game{}
	if e := mongodb.Find(context.Background(), game, 0); e != nil {
		log.Error(e)
		return nil, nil, nil, 0, e
	}
	accAmountA := &AccAmount{
		AccBet:    game.AccAmountA.AccBet,
		AccPayout: game.AccAmountA.AccPayout,
	}
	accAmountB := &AccAmount{
		AccBet:    game.AccAmountB.AccBet,
		AccPayout: game.AccAmountB.AccPayout,
	}
	accAmountPeace := &AccAmount{
		AccBet:    game.AccAmountPeace.AccBet,
		AccPayout: game.AccAmountPeace.AccPayout,
	}
	if accAmountA.AccBet > 0 {
		accAmountA.Rate = float64(accAmountA.AccPayout) / float64(accAmountA.AccBet)
	}
	if accAmountB.AccBet > 0 {
		accAmountB.Rate = float64(accAmountB.AccPayout) / float64(accAmountB.AccBet)
	}
	if accAmountPeace.AccBet > 0 {
		accAmountPeace.Rate = float64(accAmountPeace.AccPayout) / float64(accAmountPeace.AccBet)
	}
	return accAmountA, accAmountB, accAmountPeace, game.Amount, nil
}
func Data() fiber.Handler {
	return func(c *fiber.Ctx) error {
		accAmountA, accAmountB, accAmountPeace, Amount, err := GetData()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1000,
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"A":             accAmountA,
				"B":             accAmountB,
				"Peace":         accAmountPeace,
				"AvailableFund": Amount, // 可赔付库存
			},
		})
	}
}
