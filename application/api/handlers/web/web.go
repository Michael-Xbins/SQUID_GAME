package web

import (
	"application/api/presenter"
	"application/api/presenter/web"
	"application/mongodb"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/speps/go-hashids"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
	"time"
)

var Whitelist = []string{"7938939851", "2141546562", "5243341575", "7891841063", "6518048326", "6387127034", "6370621767"}

const cdkMinLength = 10

func isWhite(account string) bool {
	for _, id := range Whitelist {
		if account == id {
			return true
		}
	}
	return false
}

type cdkDetail struct {
	Cdk        string                 `json:"cdk"`
	BatchNum   int64                  `json:"batchNum"`
	Type       string                 `json:"type"`
	Rewards    []presenter.RewardInfo `json:"rewards"`
	Deadline   int64                  `json:"deadline"`
	Times      int                    `json:"times"`
	UserNums   int                    `json:"userNums"` //使用人次
	Channel    string                 `json:"channel"`
	IsRecharge bool                   `json:"isRecharge"`
	Users      []string               `json:"users"`
}

func retCdkInfo(cdk string) (*cdkDetail, error) {
	cdkInfo := &presenter.Cdk{}
	if err := mongodb.Find(context.Background(), cdkInfo, cdk); err != nil {
		return nil, err
	}
	users := make([]string, 0)
	for uid := range cdkInfo.ExchangeUid {
		users = append(users, uid)
	}
	detail := &cdkDetail{
		Cdk:        cdkInfo.Cdk,
		BatchNum:   cdkInfo.BatchNum,
		Type:       cdkInfo.Type,
		Rewards:    cdkInfo.Rewards,
		Deadline:   cdkInfo.Deadline,
		Times:      cdkInfo.Times,
		UserNums:   len(cdkInfo.ExchangeUid),
		Channel:    cdkInfo.Channel,
		IsRecharge: cdkInfo.IsRecharge,
		Users:      users,
	}
	return detail, nil
}

func GetLabel() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type statusRequest struct {
			SessionToken string `json:"sessionToken"`
		}
		var req statusRequest
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

		// 检查 UID 是否在白名单中
		if !isWhite(account) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "UID is not whitelisted",
			})
		}
		labels, err := mongodb.FindLabels()
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1004,
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"labels": labels,
			},
		})
	}
}

func AddLabel() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type statusRequest struct {
			SessionToken string `json:"sessionToken"`
			ID           int64  `json:"id"`  //标签ID
			Adv          string `json:"adv"` //标签名
		}
		var req statusRequest
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
		if req.ID < 10000 || req.Adv == "" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "params error",
			})
		}
		// 检查 UID 是否在白名单中
		if !isWhite(account) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1004,
				Message: "UID is not whitelisted",
			})
		}
		if isExist, err := mongodb.ExistAdv(req.Adv); err != nil {
			log.Error(err)
		} else if isExist {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1005,
				Message: "Adv already exists",
			})
		}

		label := &web.Label{}
		if err := mongodb.Find(context.Background(), label, req.ID); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				label.Id = req.ID
				label.Adv = req.Adv
				label.Timestamp = time.Now().UnixMilli()
				err := mongodb.Insert(context.Background(), label)
				if err != nil {
					log.Error(err)
					return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
						Code: 1006,
					})
				}
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1,
					Message: "Success",
					Result: map[string]interface{}{
						"label": label,
					},
				})
			}
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1007,
			})
		} else {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1008,
				Message: "ID already exists",
			})
		}
	}
}

func AdvUrl() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type statusRequest struct {
			SessionToken string `json:"sessionToken"`
			ID           int64  `json:"id"` //标签ID
		}
		var req statusRequest
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
		if req.ID < 10000 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1001,
				Message: "params error",
			})
		}
		// 检查 UID 是否在白名单中
		if !isWhite(account) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "UID is not whitelisted",
			})
		}

		// 生成广告链接
		payload := fmt.Sprintf("adv_%d", req.ID)                                  //广告标签ID
		targetURL := "https://t.me/" + utils.BotUsername + "?startapp=" + payload //直接启动小程序

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"links": targetURL,
			},
		})
	}
}

func GetCDK() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
			Type         string `json:"type"` // assign, single, multi
			// assign
			Account string `json:"account"` // 指定账户名
			// single, multi
			Days       int                    `json:"days"`       // 几天后截止
			UseTimes   int                    `json:"useTimes"`   // 限用次数
			IsRecharge bool                   `json:"isRecharge"` // 是否为充值后才可用
			Channel    string                 `json:"channel"`    // 生效渠道
			Num        int                    `json:"num"`        // 生成cdk数量
			Rewards    []presenter.RewardInfo `json:"rewards"`    // cdk兑换奖励列表
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
		// 检查 UID 是否在白名单中
		if !isWhite(account) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "UID is not whitelisted",
			})
		}
		if req.Type != "assign" && req.Type != "single" && req.Type != "multi" {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1004,
				Message: "type param error",
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
		for _, reward := range req.Rewards {
			// TODO:目前仅支持coin金币
			if reward.RewardId != "coin" {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1007,
					Message: "rewardId param error",
				})
			}
		}
		if req.Num < 1 || req.UseTimes < 1 || req.Days < 1 {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1008,
				Message: "param error",
			})
		}

		batchCdk := &presenter.BatchCdk{}
		// 获取下一个序列号
		nextNum, err := mongodb.GetBatchCdkNextSeq(batchCdk.TableName())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1009,
			})
		}
		var cdk []string
		var salt string
		hd := hashids.NewData()
		hd.MinLength = cdkMinLength // 设置生成的哈希ID的最小长度
		for i := 0; i < req.Num; i++ {
			var e string

			if req.Type == "assign" {
				if req.Account == "" {
					return c.Status(fiber.StatusOK).JSON(presenter.Response{
						Code:    1010,
						Message: "type param error",
					})
				}

				// 使用用户账号和当前时间戳作为盐值
				salt = fmt.Sprintf("%s-%d", req.Account, time.Now().UnixNano())
				hd.Salt = salt
				h, _ := hashids.NewWithData(hd)
				assignUid, err := strconv.Atoi(req.Account)
				if err != nil {
					log.Error(err)
					return c.Status(fiber.StatusOK).JSON(presenter.Response{
						Code:    1011,
						Message: "account param error",
					})
				}
				e, _ = h.Encode([]int{assignUid})
				log.Debugf("运营:%s生成CDK, 指定用户:%s,  CDK:%v", userInfo.Account, req.Account, e)

			} else {
				salt, _ = generateSalt(16) // 每个批次使用一个固定的盐值
				hd.Salt = salt
				h, _ := hashids.NewWithData(hd)
				id, _ := h.Encode([]int{time.Now().Nanosecond(), i, rand.Intn(1000)})
				e = id
			}

			cdkInfo := &presenter.Cdk{
				Cdk:         e,
				BatchNum:    nextNum,
				Type:        req.Type,
				Deadline:    time.Now().Add(time.Duration(req.Days) * 24 * time.Hour).UnixMilli(),
				Times:       req.UseTimes,
				Channel:     req.Channel,
				IsRecharge:  req.IsRecharge,
				Rewards:     req.Rewards,
				ExchangeUid: make(map[string]int),
				Salt:        salt,
			}
			if err := mongodb.Insert(context.Background(), cdkInfo); err != nil {
				log.Error(err)
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1012,
				})
			}
			cdk = append(cdk, e)
		}

		batchCdk.Num = nextNum
		batchCdk.CDKs = cdk
		batchCdk.SingleUid = map[string]string{}
		batchCdk.Timestamp = time.Now().UnixMilli()
		if err := mongodb.Insert(context.Background(), batchCdk); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1013,
			})
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"CDK":      cdk,
				"batchNum": nextNum,
			},
		})
	}
}

func generateSalt(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
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
		oldBalance := userInfo.Balance
		oldVoucher := userInfo.Voucher
		cdkInfo := &presenter.Cdk{}
		if err := mongodb.Find(context.Background(), cdkInfo, req.Cdk); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1005,
				Message: "Invalid CDK",
			})
		}
		if time.Now().UnixMilli() > cdkInfo.Deadline {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1006,
				Message: "CDK has expired",
			})
		}
		if cdkInfo.Channel != "" && userInfo.Channel != cdkInfo.Channel {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1007,
				Message: "Channel cannot be exchanged",
			})
		}
		if cdkInfo.IsRecharge {
			PayInfos, err := mongodb.FindAllRechargeOrder(context.Background(), account)
			if err != nil {
				log.Error(err)
			}
			if len(PayInfos) == 0 {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1008,
					Message: "Not recharged",
				})
			}
		}
		batchCdk := &presenter.BatchCdk{}
		if err := mongodb.Find(context.Background(), batchCdk, cdkInfo.BatchNum); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1009,
			})
		}

		if cdkInfo.Type == "assign" {
			hdd := hashids.NewData()
			hdd.Salt = cdkInfo.Salt
			hdd.MinLength = cdkMinLength
			hh, _ := hashids.NewWithData(hdd)
			assignUid, err := hh.DecodeWithError(req.Cdk)
			if err != nil {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1010,
					Message: "Invalid CDK",
				})
			}
			if strconv.Itoa(assignUid[0]) != account {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1011,
					Message: "User not assigned",
				})
			}
			if _, exists := cdkInfo.ExchangeUid[account]; !exists {
				cdkInfo.ExchangeUid[account] = 0
			}
		} else if cdkInfo.Type == "single" {
			value, exists := batchCdk.SingleUid[account]
			if exists && value != "" && value != req.Cdk {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1012,
					Message: "Already occupied",
				})
			} else {
				if cdkInfo.ExchangeUid == nil {
					cdkInfo.ExchangeUid[account] = 0
				} else {
					for k := range cdkInfo.ExchangeUid {
						if k != account {
							return c.Status(fiber.StatusOK).JSON(presenter.Response{
								Code:    1013,
								Message: "Already occupied",
							})
						}
					}
				}
				batchCdk.SingleUid[account] = req.Cdk
			}

		} else if cdkInfo.Type == "multi" {
			if _, exists := cdkInfo.ExchangeUid[account]; !exists {
				cdkInfo.ExchangeUid[account] = 0
			}
		}
		if cdkInfo.ExchangeUid[account] >= cdkInfo.Times {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1014,
				Message: "Exceeding the usage limit",
			})
		}
		cdkInfo.ExchangeUid[account] += 1

		for _, reward := range cdkInfo.Rewards {
			if reward.RewardId == "coin" {
				mongodb.AddAmount(userInfo, reward.RewardNum)
				// coinFlow埋点
				log.InfoJson("金币入口",
					zap.String("Account", userInfo.Account),
					zap.String("ActionType", log.Flow),
					zap.String("FlowType", log.CoinFlow),
					zap.String("From", log.CDK),
					zap.String("Flag", log.FlagIn),
					zap.Int64("Amount", reward.RewardNum), //兑换了
					zap.Int64("Old", oldBalance),          //旧游戏币
					zap.Int64("New", userInfo.Balance),    //新游戏币
					zap.Int64("CreatedAt", time.Now().UnixMilli()),
				)
			} else if reward.RewardId == "voucher" {
				mongodb.AddVoucher(userInfo, reward.RewardNum)
				// VoucherFlow埋点
				log.InfoJson("凭证入口",
					zap.String("Account", userInfo.Account),
					zap.String("ActionType", log.Flow),
					zap.String("FlowType", log.VoucherFlow),
					zap.String("From", log.CDK),
					zap.String("Flag", log.FlagIn),
					zap.Int64("Amount", reward.RewardNum), //兑换了
					zap.Int64("Old", oldVoucher),          //旧券
					zap.Int64("New", userInfo.Voucher),    //新券
					zap.Int64("CreatedAt", time.Now().UnixMilli()),
				)
			}
		}
		if err := mongodb.Update(context.Background(), userInfo, nil); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1015,
			})
		}
		if err := mongodb.Update(context.Background(), cdkInfo, nil); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1016,
			})
		}
		if err := mongodb.Update(context.Background(), batchCdk, nil); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1017,
			})
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"rewards": cdkInfo.Rewards,
				"balance": userInfo.Balance,
				"voucher": userInfo.Voucher,
			},
		})
	}
}

func QueryCdk() fiber.Handler {
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
		// 检查 UID 是否在白名单中
		if !isWhite(account) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "UID is not whitelisted",
			})
		}
		userInfo := &presenter.UserInfo{}
		err = mongodb.Find(context.Background(), userInfo, account)
		if err != nil {
			log.Error(err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1004,
					Message: "Invalid CDK",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1005,
			})
		}
		cdkInfo, err := retCdkInfo(req.Cdk)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1006,
			})
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"info": cdkInfo,
			},
		})
	}
}

func QueryBatchCdk() fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			SessionToken string `json:"sessionToken"`
			BatchNum     int64  `json:"batchNum"`
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
		// 检查 UID 是否在白名单中
		if !isWhite(account) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "UID is not whitelisted",
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
		batchCdk := &presenter.BatchCdk{}
		if err := mongodb.Find(context.Background(), batchCdk, req.BatchNum); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1006,
			})
		}
		cdkInfos := make([]*cdkDetail, 0)
		totalUserNums := 0
		for _, cdk := range batchCdk.CDKs {
			cdkInfo, err := retCdkInfo(cdk)
			if err != nil {
				log.Error(err)
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1007,
				})
			}
			cdkInfos = append(cdkInfos, cdkInfo)
			totalUserNums += cdkInfo.UserNums
		}
		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"infos":         cdkInfos,
				"totalUserNums": totalUserNums,
			},
		})
	}
}

func HistoryCdk() fiber.Handler {
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
		// 检查 UID 是否在白名单中
		if !isWhite(account) {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1003,
				Message: "UID is not whitelisted",
			})
		}
		batchCdks, err := mongodb.FindAllBatchCdk(context.Background())
		if err != nil {
			log.Error("FindAllBatchCdk error: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
				Code: 1004,
			})
		}

		type historyCdk struct {
			BatchNum   int64                  `json:"batchNum"`
			Type       string                 `json:"type"`
			Reward     []presenter.RewardInfo `json:"reward"`
			Deadline   int64                  `json:"deadline"`
			Times      int                    `json:"times"`
			Nums       int                    `json:"nums"` //本批次生成CDK数量
			Channel    string                 `json:"channel"`
			IsRecharge bool                   `json:"isRecharge"`
			Timestamp  int64                  `json:"timestamp"`
		}

		var history historyCdk
		histories := make([]historyCdk, 0)
		for _, info := range batchCdks {
			cdkInfo, err := retCdkInfo(info.CDKs[0])
			if err != nil {
				log.Error(err)
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code: 1005,
				})
			}
			history.BatchNum = info.Num
			history.Type = cdkInfo.Type
			history.Reward = cdkInfo.Rewards
			history.Deadline = cdkInfo.Deadline
			history.Times = cdkInfo.Times
			history.Nums = len(info.CDKs)
			history.Channel = cdkInfo.Channel
			history.IsRecharge = cdkInfo.IsRecharge
			history.Timestamp = info.Timestamp
			histories = append(histories, history)
		}

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code: 1,
			Result: map[string]interface{}{
				"history": histories,
			},
		})
	}
}
