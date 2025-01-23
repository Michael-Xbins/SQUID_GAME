package auction

import (
	"application/api/presenter"
	"application/api/presenter/auction"
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
	"time"
)

func Sync(f func(*fiber.Ctx) error) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var errChan = make(chan error)
		select {
		case fchan <- func() {
			defer close(errChan)
			errChan <- f(c)
		}:
		default:
			log.Error("队列阻塞")
			return errors.New("queue blocked")
		}
		return <-errChan
	}
}

func checkOpenTime() error {
	starttime := int64(utils.LubanTables.TBApp.Get("starttime").NumInt)
	endtime := int64(utils.LubanTables.TBApp.Get("endtime").NumInt)
	now := time.Now()
	secondsSinceMidnight := int64(now.Hour()*3600 + now.Minute()*60 + now.Second())
	if secondsSinceMidnight >= endtime || secondsSinceMidnight < starttime {
		return errors.New("the market has been closed")
	}
	return nil
}

func Buy(c *fiber.Ctx) error {
	type statusRequest struct {
		SessionToken string `json:"sessionToken"`
		Count        int64  `json:"count"`
		Price        int64  `json:"price"`
	}
	if err := checkOpenTime(); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1009,
			Message: err.Error(),
		})
	}
	var status statusRequest
	if err := c.BodyParser(&status); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1000,
		})
	}
	if status.SessionToken == "" || status.Count < 1 || status.Price < 1 || status.Count > 1000000 {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1001,
		})
	}
	tradeSum := int64(utils.LubanTables.TBApp.Get("trade_sum").NumInt)
	limit := game.ClosePrice * tradeSum / 1000
	if status.Price < game.ClosePrice-limit || status.Price > game.ClosePrice+limit {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1002,
			Message: "Exceeding the purchase limit",
		})
	}

	totalAmount := status.Count * status.Price
	account, err := redis.GetAccountBySessionToken(status.SessionToken)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1003,
		})
	}
	userInfo := &presenter.UserInfo{}
	err = mongodb.Find(context.Background(), userInfo, account)
	if err != nil {
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
	if userInfo.Balance < totalAmount {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1006,
		})
	}
	userInfo.Balance -= totalAmount
	buyOrder := &auction.BuyOrderList{
		Id:             uuid.NewString(),
		Timestamp:      time.Now().UnixMilli(),
		Account:        account,
		TotalCount:     status.Count,
		CompletedCount: 0,
		Price:          status.Price,
	}
	err = mongodb.Update(context.Background(), userInfo, nil)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1007,
		})
	}
	err = mongodb.Insert(context.Background(), buyOrder)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1008,
		})
	}
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_BuyOrderNotify_,
		Message: &pb.ClientResponse_BuyOrderNotify{BuyOrderNotify: &pb.BuyOrderNotify{
			BuyOrderInfo: &pb.BuyOrderInfo{
				Id:    buyOrder.Id,
				Count: buyOrder.TotalCount,
				Price: buyOrder.Price,
			},
		}},
	})

	return c.Status(fiber.StatusOK).JSON(presenter.Response{
		Code: 1,
		Result: map[string]interface{}{
			"balance": userInfo.Balance,
		},
	})
}

func Sell(c *fiber.Ctx) error {
	type statusRequest struct {
		SessionToken string `json:"sessionToken"`
		Count        int64  `json:"count"`
		Price        int64  `json:"price"`
	}
	if err := checkOpenTime(); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1009,
			Message: err.Error(),
		})
	}
	var status statusRequest
	if err := c.BodyParser(&status); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1000,
		})
	}
	if status.SessionToken == "" || status.Count < 1 || status.Price < 1 || status.Count > 1000000 {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1001,
		})
	}
	tradeSum := int64(utils.LubanTables.TBApp.Get("trade_sum").NumInt)
	limit := game.ClosePrice * tradeSum / 1000
	if status.Price < game.ClosePrice-limit || status.Price > game.ClosePrice+limit {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1002,
			Message: "Exceeding the sales limit",
		})
	}

	account, err := redis.GetAccountBySessionToken(status.SessionToken)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1003,
		})
	}
	userInfo := &presenter.UserInfo{}
	err = mongodb.Find(context.Background(), userInfo, account)
	if err != nil {
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
	if userInfo.Voucher < status.Count {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1006,
		})
	}
	userInfo.Voucher -= status.Count
	sellOrder := &auction.SellOrderList{
		Id:             uuid.NewString(),
		Timestamp:      time.Now().UnixMilli(),
		Account:        account,
		TotalCount:     status.Count,
		CompletedCount: 0,
		Price:          status.Price,
	}
	err = mongodb.Insert(context.Background(), sellOrder)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1007,
		})
	}
	err = mongodb.Update(context.Background(), userInfo, nil)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1008,
		})
	}
	session.S2AllMessage(&pb.ClientResponse{
		Type: pb.MessageType_SellOrderNotify_,
		Message: &pb.ClientResponse_SellOrderNotify{SellOrderNotify: &pb.SellOrderNotify{
			SellOrderInfo: &pb.SellOrderInfo{
				Id:    sellOrder.Id,
				Count: sellOrder.TotalCount,
				Price: sellOrder.Price,
			},
		}},
	})

	return c.Status(fiber.StatusOK).JSON(presenter.Response{
		Code: 1,
		Result: map[string]interface{}{
			"voucher": userInfo.Voucher,
		},
	})
}

func TradeList(c *fiber.Ctx) error {
	type statusRequest struct {
		SessionToken string `json:"sessionToken"`
	}
	var status statusRequest
	if err := c.BodyParser(&status); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1000,
		})
	}

	account, err := redis.GetAccountBySessionToken(status.SessionToken)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1002,
		})
	}
	buyOrderList, err := mongodb.GetBuyOrderListByAccount(context.Background(), account)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1004,
		})
	}
	sellOrderList, err := mongodb.GetSellOrderListByAccount(context.Background(), account)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1005,
		})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.Response{
		Code: 1,
		Result: map[string]interface{}{
			"buyOrderList":  buyOrderList,
			"sellOrderList": sellOrderList,
		},
	})
}

func Cancel(c *fiber.Ctx) error {
	type statusRequest struct {
		SessionToken string `json:"sessionToken"`
		Id           string `json:"id"`
		Type         string `json:"type"`
	}
	var status statusRequest
	if err := c.BodyParser(&status); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1000,
		})
	}
	if status.SessionToken == "" {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1001,
		})
	}
	account, err := redis.GetAccountBySessionToken(status.SessionToken)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1002,
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
	if status.Type == "buy" {
		buyOrderList := &auction.BuyOrderList{}
		err = mongodb.Find(context.Background(), buyOrderList, status.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1007,
			})
		}
		err = mongodb.Delete(context.Background(), buyOrderList)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1008,
			})
		}
		cnt := buyOrderList.TotalCount - buyOrderList.CompletedCount
		if cnt > 0 {
			restitutionAmount := cnt * buyOrderList.Price
			userInfo.Balance += restitutionAmount
			err = mongodb.Update(context.Background(), userInfo, nil)
			if err != nil {
				log.Error(err)
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1009,
				})
			}
		}

	} else {
		sellOrderList := &auction.SellOrderList{}
		err = mongodb.Find(context.Background(), sellOrderList, status.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1010,
			})
		}
		err = mongodb.Delete(context.Background(), sellOrderList)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1011,
			})
		}
		cnt := sellOrderList.TotalCount - sellOrderList.CompletedCount
		userInfo.Voucher += cnt
		err = mongodb.Update(context.Background(), userInfo, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1012,
			})
		}
	}
	orderChanged = true
	return c.Status(fiber.StatusOK).JSON(presenter.Response{
		Code: 1,
		Result: map[string]interface{}{
			"balance": userInfo.Balance,
			"voucher": userInfo.Voucher,
		},
	})
}

func History(c *fiber.Ctx) error {
	type statusRequest struct {
		SessionToken string `json:"sessionToken"`
		Day          int32  `json:"day"`
	}
	var status statusRequest
	if err := c.BodyParser(&status); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1000,
		})
	}
	if status.SessionToken == "" {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1001,
		})
	}
	account, err := redis.GetAccountBySessionToken(status.SessionToken)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1002,
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
	history, err := mongodb.GetOrderHistoryByDate(context.Background(), account, status.Day)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
			Code: 1007,
		})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.Response{
		Code: 1,
		Result: map[string]interface{}{
			"history": history,
		},
	})
}

func Login(c *fiber.Ctx) error {
	type statusRequest struct {
		SessionToken string `json:"sessionToken"`
	}
	var status statusRequest
	if err := c.BodyParser(&status); err != nil {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1000,
		})
	}
	if status.SessionToken == "" {
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1001,
		})
	}
	account, err := redis.GetAccountBySessionToken(status.SessionToken)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1002,
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

	var sellListTop5, _ = mongodb.GetTop5SellOrderList(context.Background())
	var buyListTop5, _ = mongodb.GetTop5BuyOrderList(context.Background())
	return c.Status(fiber.StatusOK).JSON(presenter.Response{
		Code: 1,
		Result: map[string]interface{}{
			"tradeRecordList": game.TradeRecordList,
			"buyRecordList":   buyListTop5,
			"sellRecordList":  sellListTop5,
			"closePrice":      game.ClosePrice,
			"serverTime":      time.Now().UnixMilli(),
		},
	})
}
