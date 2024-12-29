package handlers

import (
	"application/api/presenter"
	rechargedb "application/api/presenter/recharge"
	"application/mongodb"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/recharge"
	"application/redis"
	"application/session"
	"application/wallet"
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/speps/go-hashids"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func GetInvitesHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var inviteReq presenter.InviteReq
		if err := c.BodyParser(&inviteReq); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		sessionToken := inviteReq.SessionToken
		if sessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
			})
		}
		account, err := redis.GetAccountBySessionToken(sessionToken)
		if err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "sessionToken error",
			})
		}

		//从db取出user数据至内存
		userInfo := &presenter.UserInfo{}
		err = mongodb.Find(context.Background(), userInfo, account)
		if err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code:    1003,
					Message: "No invited_account found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1004,
				Message: "Failed to retrieve invites",
			})
		}
		mongodb.UpdateUserinfo(userInfo)
		today := time.Now().Format("2006-01-02")
		DailyInvites, _ := redis.GetInvitesFromToday(account, today)
		downLinesInfos := make([]presenter.DownLineInfo, 0)
		for _, downLine := range userInfo.DownLines {
			downDownLinesUsdtRecharge := int64(0)
			dailyDownDownLinesUsdtRecharge := int64(0)
			downLineUser := &presenter.UserInfo{}

			err := mongodb.Find(context.Background(), downLineUser, downLine)
			if err != nil {
				log.Errorf("downLine: %s, err: %v", downLine, err)
			}
			mongodb.UpdateUserinfo(downLineUser)

			for _, downDownLine := range downLineUser.DownLines {
				downDownLineUser := &presenter.UserInfo{}
				err := mongodb.Find(context.Background(), downDownLineUser, downDownLine)
				if err != nil {
					log.Errorf("downDownLine: %s, err: %v", downDownLine, err)
				}
				dailyDownDownLinesUsdtRecharge += downDownLineUser.UsdtRecharge.DailyRecharge
				downDownLinesUsdtRecharge += downDownLineUser.UsdtRecharge.TotalRecharge
			}

			downLineInfo := presenter.DownLineInfo{
				ID:                             downLineUser.Account,
				Name:                           downLineUser.Nickname,
				UsdtRecharge:                   downLineUser.UsdtRecharge,
				UsdtAgent:                      downLineUser.UsdtAgent,
				DownDownLinesUsdtDailyRecharge: dailyDownDownLinesUsdtRecharge, // 仅downLineID用户 所有下线今日的充值流水
				DownDownLinesUsdtRecharge:      downDownLinesUsdtRecharge,      // 仅downLineID用户 所有下线总充值流水
			}
			downLinesInfos = append(downLinesInfos, downLineInfo)
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"today_invited_times":             len(DailyInvites),
				"all_invited_times":               len(userInfo.DownLines),
				"today_invited_accounts":          DailyInvites,
				"all_invited_accounts":            userInfo.DownLines,
				"downLine_dailyRecharge_accounts": len(userInfo.UsdtRecharge.DownLineDailyRecharge),
				"downLine_totalRecharge_accounts": len(userInfo.UsdtRecharge.DownLineTotalRecharge),
				"completed_tasks":                 userInfo.CompletedTasks,
				"usdt_agent":                      userInfo.UsdtAgent,
				"downLinesInfos":                  downLinesInfos,
			},
		})
	}
}

func ClaimInviteReward() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var claimReq presenter.ClaimReq
		if err := c.BodyParser(&claimReq); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		sessionToken := claimReq.SessionToken
		if sessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
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

		//从db取出user数据至内存
		userInfo := &presenter.UserInfo{}
		err = mongodb.Find(context.Background(), userInfo, account)
		if err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code:    1003,
					Message: "No invited_account found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1004,
				Message: "Failed to retrieve invites",
			})
		}
		taskInfo := utils.LubanTables.TBApp.Get(claimReq.TaskId)
		if taskInfo == nil {
			return c.Status(fiber.StatusBadRequest).JSON(presenter.Response{
				Code:    1005,
				Message: "TaskId not exist",
			})
		}
		if len(userInfo.DownLines) < int(taskInfo.Target) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1006,
				Message: "invites not enough",
			})
		}

		// 检查任务是否已经完成
		completed := false
		for _, taskId := range userInfo.CompletedTasks {
			if taskId == claimReq.TaskId {
				completed = true
				break
			}
		}
		if completed {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1007,
				Message: "already claimed",
			})
		}

		// 添加任务完成列表, 账户增加代币
		userInfo.CompletedTasks = append(userInfo.CompletedTasks, claimReq.TaskId)
		mongodb.AddAmount(userInfo, int64(taskInfo.NumInt))
		err = mongodb.Update(context.Background(), userInfo, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1008,
				Message: "Failed to update task status",
			})
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Reward claimed successfully",
			Result: map[string]interface{}{
				"bonus":   taskInfo.NumInt,
				"balance": userInfo.Balance,
			},
		})
	}
}

func GetUserInfo() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var inviteReq presenter.InviteReq
		if err := c.BodyParser(&inviteReq); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		sessionToken := inviteReq.SessionToken
		if sessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken is required",
			})
		}
		account, err := redis.GetAccountBySessionToken(sessionToken)
		if err != nil {
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
					Code:    1003,
					Message: "No invited_account found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1004,
				Message: "Failed to retrieve invites",
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"Nickname": userInfo.Nickname,
				"ID":       userInfo.Account,
			},
		})
	}
}

func SetName() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type setNameRequest struct {
			SessionToken string `json:"sessionToken"`
			Name         string `json:"name"`
		}
		var request setNameRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if request.SessionToken == "" || request.Name == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "sessionToken or name is required",
			})
		}
		account, err := redis.GetAccountBySessionToken(request.SessionToken)
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
					Code:    1003,
					Message: "No account found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1004,
				Message: "Failed to SetName",
			})
		}
		if userInfo.Nickname == request.Name {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1005,
				Message: "Same as original name",
			})
		}

		userInfo.Nickname = request.Name
		err = mongodb.Update(context.Background(), userInfo, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code:    1006,
				Message: "Failed to set name",
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"NewNickname": userInfo.Nickname,
			},
		})
	}
}

func ClaimAgent() fiber.Handler {
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

		if userInfo.UsdtAgent.Unclaimed <= 0 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1005,
				Message: "The commission you can receive is 0",
			})
		}

		mongodb.TransferUsdtAgent(userInfo)
		err = mongodb.Update(context.Background(), userInfo, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1006,
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"USDT":         userInfo.USDT,                // 玩家usdt余额
				"NewUnclaimed": userInfo.UsdtAgent.Unclaimed, // 未领取佣金
				"NewClaimed":   userInfo.UsdtAgent.Claimed,   // 已领取佣金
			},
		})
	}
}

func GetCDK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
			Type         int    `json:"type"`
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
		if req.Type != 1 && req.Type != 2 && req.Type != 3 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "type param error",
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

		hd := hashids.NewData()
		// 使用用户账号和当前时间戳作为盐值
		salt := fmt.Sprintf("%s-%d", userInfo.Account, time.Now().UnixMilli())
		hd.Salt = salt
		hd.MinLength = 30
		h, _ := hashids.NewWithData(hd)
		e, _ := h.Encode([]int{req.Type}) // 对应 exchange1,2,3
		log.Debugf("account: %v, exchange%v CDK: %v", userInfo.Account, req.Type, e)
		key := fmt.Sprintf("%s%d", "exchange", req.Type)
		exchangeAmount := int64(utils.LubanTables.TBApp.Get(key).NumInt)
		cdkInfo := &presenter.CdkInfo{
			Cdk:            e,
			Salt:           salt,
			ExchangeID:     req.Type,
			ExchangeAmount: exchangeAmount,
			Received:       false,
		}
		if err := mongodb.Insert(context.Background(), cdkInfo); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1006,
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"CDK": e,
			},
		})
	}
}

func ExchangeCDK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
			Cdk          string `json:"cdk"`
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
		cdkInfo := &presenter.CdkInfo{}
		if err := mongodb.Find(context.Background(), cdkInfo, req.Cdk); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1005,
				Message: "Invalid CDK",
			})
		}

		hdd := hashids.NewData()
		hdd.Salt = cdkInfo.Salt
		hdd.MinLength = 30
		hh, _ := hashids.NewWithData(hdd)
		numbers, err := hh.DecodeWithError(req.Cdk)
		if err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1006,
				Message: "Invalid CDK",
			})
		}

		if cdkInfo.Received {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1007,
				Message: "CDK has already been used",
			})
		}
		log.Debugf("exchangeID: %v", numbers[0])

		// 兑换代币至余额
		key := fmt.Sprintf("%s%d", "exchange", numbers[0])
		exchangeAmount := int64(utils.LubanTables.TBApp.Get(key).NumInt)
		mongodb.AddAmount(userInfo, exchangeAmount)
		err = mongodb.Update(context.Background(), userInfo, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1008,
			})
		}

		cdkInfo.Received = true
		if err := mongodb.Update(context.Background(), cdkInfo, nil); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1009,
			})
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"ExchangeAmount": exchangeAmount,
				"Bonus":          userInfo.Balance,
			},
		})
	}
}

func MyService() fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "https://t.me/" + utils.BotUsername
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"TargetURL": targetURL,
			},
		})
	}
}

func ClaimWelfare() fiber.Handler {
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
		mongodb.UpdateUserinfo(userInfo)
		poorLim := int64(utils.LubanTables.TBApp.Get("poor_lim").NumInt)
		if userInfo.Balance >= poorLim {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1005,
				Message: "Trigger conditions not met",
			})
		}
		if userInfo.Welfare.Times < 1 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1006,
				Message: "times not enough",
			})
		}
		poorNum := int64(utils.LubanTables.TBApp.Get("poor_num").NumInt)
		userInfo.Welfare.Times -= 1
		userInfo.Welfare.TotalAmount += poorNum
		mongodb.AddAmount(userInfo, poorNum)
		err = mongodb.Update(context.Background(), userInfo, nil)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1007,
			})
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"welfare":    poorNum,
				"times":      userInfo.Welfare.Times,
				"newBalance": userInfo.Balance,
			},
		})
	}
}

func GetWelfare() fiber.Handler {
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
		mongodb.UpdateUserinfo(userInfo)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"times":   userInfo.Welfare.Times,
				"balance": userInfo.Balance,
			},
		})
	}
}

func GetAsset() fiber.Handler {
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
		mongodb.UpdateUserinfo(userInfo)
		fund := &rechargedb.Fund{}
		_ = mongodb.Find(context.Background(), fund, 0)
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"balance":         userInfo.Balance,
				"usdt":            userInfo.USDT,
				"voucher":         userInfo.Voucher,
				"withdrawnAmount": fund.WithdrawnAmount,
			},
		})
	}
}

func Recharge() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
			Network      string `json:"network"`
		}
		var req request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if req.Network != utils.NetworkTRC20 && req.Network != utils.NetworkTON {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "params error",
			})
		}
		if req.SessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "sessionToken is required",
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

		var address string
		if len(userInfo.Address) > 0 {
			address = userInfo.Address
		} else {
			address, err = recharge.AssignWalletToUser(context.Background(), req.Network, userInfo)
			if err != nil {
				log.Error("AssignWalletToUser error: ", err)
			}
		}
		qrCode, err := recharge.GenerateQRCode(address, req.Network) // 生成支付二维码
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1006,
			})
		}

		// 启动支付状态轮询
		utils.SafeGo(func() {
			startTimeMs := time.Now().UnixMilli()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) // 设置10分钟的上下文超时, 轮询协程将停止
			defer cancel()
			paymentSuccess := StartPaymentPolling(ctx, address, req.Network, userInfo.Account, startTimeMs)
			if paymentSuccess {
				log.Infof("支付成功, 玩家%s,地址:%s, 游戏币兑换处理完毕", account, address)
				newUserInfo := &presenter.UserInfo{}
				_ = mongodb.Find(context.Background(), newUserInfo, account)
				session.S2CMessage(newUserInfo.Account, &pb.ClientResponse{
					Type: pb.MessageType_RechargeNotify_,
					Message: &pb.ClientResponse_RechargeNotify{RechargeNotify: &pb.RechargeNotify{
						Address: address,
						Balance: newUserInfo.Balance,
						Usdt:    newUserInfo.USDT,
						Voucher: newUserInfo.Voucher,
					}},
				})

			} else {
				log.Infof("已超时, 玩家%s,地址:%s, 未查到交易信息", account, address)
			}
		})

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"address": address,
				"qrCode":  "data:image/png;base64," + qrCode,
			},
		})
	}
}

// StartPaymentPolling 开始轮询支付状态
func StartPaymentPolling(ctx context.Context, address string, network string, userId string, startTimeMs int64) bool {
	ticker := time.NewTicker(30 * time.Second) // 30秒检查一次
	defer ticker.Stop()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute) // 5分钟超时
	defer cancel()
	for {
		select {
		case <-timeoutCtx.Done():
			// 在超时后再次检查支付状态
			trc20Results, err := recharge.CheckTRC20Status(ctx, address, network, startTimeMs)
			if err != nil {
				return false
			}
			hasPay := processTransactions(ctx, userId, network, address, trc20Results)
			return hasPay
		case <-ticker.C:
			trc20Results, err := recharge.CheckTRC20Status(ctx, address, network, startTimeMs)
			if err != nil {
				continue
			}
			hasPay := processTransactions(ctx, userId, network, address, trc20Results)
			if hasPay {
				return true
			}
		}
	}
}
func processTransactions(ctx context.Context, userId string, network string, address string, transactions []recharge.Trc20Result) bool {
	lockKey := redis.RechargePrefix + address
	if err := redis.TryLock(ctx, lockKey, 10); err != nil {
		log.Errorf("Unable to acquire lock for address %s: %v", address, err)
		return false
	}
	defer func() {
		if err := redis.UnLock(ctx, lockKey); err != nil {
			log.Error("Failed to release lock: ", err)
		}
	}()

	hasPay := false
	for _, tx := range transactions {
		// 有未处理充值, 创建订单
		if !mongodb.IsTransactionProcessed(ctx, tx.TransactionID) {
			// 兑换游戏币
			retGameCoins, retVoucher, err := mongodb.RechargeUsdt(userId, tx.Value, tx.TokenInfo.Decimals)
			if err != nil {
				log.Error("RechargeUsdt error: ", err)
			}
			payInfo := &rechargedb.PayInfo{
				Hash:           tx.TransactionID,
				Account:        userId,
				Network:        network,
				Address:        tx.To,
				Value:          tx.Value,
				Symbol:         tx.TokenInfo.Symbol,
				Decimals:       tx.TokenInfo.Decimals,
				Name:           tx.TokenInfo.Name,
				GameCoins:      retGameCoins,
				Voucher:        retVoucher,
				BlockTimestamp: tx.BlockTimestamp,
				CreatedAt:      time.Now().UnixMilli(),
			}
			if err := mongodb.Insert(context.Background(), payInfo); err != nil {
				log.Error("Failed to create order: ", err)
			}
			hasPay = true
		}
	}
	return hasPay
}

func Withdrawal() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
			Address      string `json:"address"`
			Network      string `json:"network"`
			Amount       int64  `json:"amount"`
		}
		var req request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}
		if req.Network != utils.NetworkTRC20 && req.Network != utils.NetworkTON {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "params error",
			})
		}
		if req.SessionToken == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "sessionToken is required",
			})
		}
		//最小提现额度
		cashMin := int64(utils.LubanTables.TBApp.Get("cash_min").NumInt)
		if req.Amount < cashMin {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "The minimum withdrawal amount has not been reached",
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
		if !mongodb.CheckUSDT(userInfo, req.Amount) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1007,
				Message: "USDT not enough",
			})
		}
		fund := &rechargedb.Fund{}
		if err := mongodb.Find(context.Background(), fund, 0); err != nil {
			log.Error(err)
			return err
		}
		mongodb.DecrUSDT(userInfo, req.Amount)

		const scaleFactor = int64(1000)
		withdrawalRate := int64(utils.LubanTables.TBApp.Get("withdrawal").NumInt)
		withdrawalAmount := req.Amount * withdrawalRate / scaleFactor //手续费,千分比
		// db记录手续费,累计提现额
		mongodb.AddWithdraw(fund, req.Amount, withdrawalAmount)
		// 计算实际提现金额（美分转美元）
		realAmount := float64(req.Amount-withdrawalAmount) / 100.0
		// 转换为字符串，保留两位小数
		realAmountStr := fmt.Sprintf("%.2f", realAmount)
		orderId, err := wallet.Withdrawal(req.Address, req.Network, realAmountStr)
		if err != nil {
			log.Error("Withdrawal error: ", err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1008,
				Message: "params error",
			})
		}

		if err = mongodb.Update(context.Background(), fund, nil); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1009,
			})
		}
		if err = mongodb.Update(context.Background(), userInfo, nil); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1010,
			})
		}
		hash := ""
		ret := wallet.WithdrawHistoryById(orderId)
		if ret != "" && ret != wallet.ErrNoTransactionFound && ret != wallet.ErrFailedToRetrieve {
			hash = ret
		}

		// 记录提现订单
		withdrawInfo := &rechargedb.WithDrawInfo{
			OrderId:    orderId,
			Account:    userInfo.Account,
			TxHash:     hash,
			Network:    req.Network,
			Address:    req.Address,
			Amount:     req.Amount,
			ServiceFee: withdrawalAmount,
			RealAmount: req.Amount - withdrawalAmount,
			CreatedAt:  time.Now().UnixMilli(),
		}
		if err := mongodb.Insert(context.Background(), withdrawInfo); err != nil {
			log.Error("Failed to create order: ", err)
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"balance":         userInfo.Balance,
				"usdt":            userInfo.USDT,
				"voucher":         userInfo.Voucher,
				"orderId":         orderId,
				"withdrawnAmount": fund.WithdrawnAmount,
			},
		})
	}
}

func HistoryRecord() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
			Type         int    `json:"type"` // 1:充值记录; 2:提现记录; 3:闪兑记录
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

		if req.Type == 1 {
			PayInfos, err := mongodb.FindAllRechargeOrder(context.Background(), userInfo.Account)
			if err != nil {
				log.Error(err)
			}
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code: 1,
				Result: map[string]interface{}{
					"rechargeRecord": PayInfos,
				},
			})
		} else if req.Type == 2 {
			withDrawInfos, err := mongodb.FindAllWithdrawOrder(context.Background(), userInfo.Account)
			if err != nil {
				log.Error(err)
			}
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code: 1,
				Result: map[string]interface{}{
					"withdrawRecord": withDrawInfos,
				},
			})
		} else if req.Type == 3 {
			exchangeInfos, err := mongodb.FindAllExchangeOrder(context.Background(), userInfo.Account)
			if err != nil {
				log.Error(err)
			}
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code: 1,
				Result: map[string]interface{}{
					"exchangeRecord": exchangeInfos,
				},
			})
		} else {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1005,
				Message: "params error",
			})
		}
	}
}
