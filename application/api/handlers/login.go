package handlers

import (
	"application/api/presenter"
	"application/api/presenter/squid"
	"application/mongodb"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	tginit "github.com/telegram-mini-apps/init-data-golang"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func checkIfNewUser(account string) (bool, error) {
	userInfo := &presenter.UserInfo{}
	err := mongodb.Find(context.Background(), userInfo, account)
	if errors.Is(err, mongo.ErrNoDocuments) && account != "" && account != "0" {
		return true, nil
	} else {
		return false, err
	}
}

func LoginHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var loginReq presenter.AppLoginReq
		if err := c.BodyParser(&loginReq); err != nil {
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1000,
				Message: "Invalid request body",
			})
		}

		if loginReq.Type == "app" {
			// 校验telegram initData
			err := tginit.Validate(loginReq.WebAppInitData, utils.TelegramToken, 0)
			if err != nil {
				log.Error("Validate telegram_WebAppInitData error: ", err)
				return c.Status(fiber.StatusOK).JSON(presenter.Response{
					Code:    1001,
					Message: "Authorization failed",
				})
			}
		}

		webAppInitData, err := tginit.Parse(loginReq.WebAppInitData)
		if err != nil {
			log.Error(err)
		}
		account := strconv.FormatInt(webAppInitData.User.ID, 10)

		// 检查是否是新用户
		isNewUser, err := checkIfNewUser(account)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusOK).JSON(presenter.Response{
				Code:    1002,
				Message: "Failed to check user status",
			})
		}

		// 提取邀请人ID
		startParam := webAppInitData.StartParam
		parts := strings.Split(startParam, "_")
		var inviter string
		if len(parts) > 1 && parts[0] == "code" && parts[1] != account {
			inviter = parts[1]
		}
		inviterNewUser, _ := checkIfNewUser(inviter)

		//重新生成sessionToken
		sessionToken := uuid.New().String()
		if isNewUser {
			var upLine string
			if !inviterNewUser {
				upLine = inviter
			}
			times := utils.LubanTables.TBApp.Get("poor_count").NumInt
			userInfo := &presenter.UserInfo{
				Account:        account,
				Nickname:       webAppInitData.User.FirstName + " " + webAppInitData.User.LastName,
				SessionToken:   sessionToken,
				Balance:        utils.InitBalance,
				USDT:           utils.InitUSDT,
				Voucher:        utils.InitVoucher,
				UpLine:         upLine,
				DownLines:      make([]string, 0),
				CompletedTasks: make([]string, 0),
				UsdtRecharge:   presenter.UsdtRechargeDetail{DownLineDailyRecharge: make(map[string]bool), DownLineTotalRecharge: make(map[string]bool)},
				Squid:          presenter.Squid{RoundId: 1, BetPricesPerRound: make([]int64, squid.TotalRounds)},
				Welfare:        presenter.Welfare{LastDate: time.Now().Format("2006-01-02"), Times: times},
			}
			err := mongodb.Insert(context.Background(), userInfo)
			if err != nil {
				log.Error(err)
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code:    1003,
					Message: "Failed to initialize user",
				})
			}
		} else {
			userInfo := &presenter.UserInfo{}
			err = mongodb.Find(context.Background(), userInfo, account)
			if err != nil {
				log.Error(err)
				return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
					Code:    1004,
					Message: "Failed to retrieve invites",
				})
			}
			userInfo.SessionToken = sessionToken
			if e := mongodb.Update(context.Background(), userInfo, nil); e != nil {
				return e
			}
		}

		// 处理邀请人信息, 邀请对象必须是新用户
		if inviter != "" && isNewUser && inviter != account {
			if !inviterNewUser {
				inviterInfo := &presenter.UserInfo{}
				err = mongodb.Find(context.Background(), inviterInfo, inviter)
				if err != nil {
					log.Error(err)
					if errors.Is(err, mongo.ErrNoDocuments) {
						return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
							Code:    1005,
							Message: "No invited_account found",
						})
					}
					return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
						Code:    1006,
						Message: "Failed to retrieve invites",
					})
				}
				today := time.Now().Format("2006-01-02")
				// 更新邀请人信息
				if err := redis.AddInviteToRedis(inviter, account, today); err != nil {
					log.Error(err)
					return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
						Code:    1007,
						Message: err.Error(),
					})
				}
				update := bson.M{
					"$push": bson.M{
						"downLines": account,
					},
				}
				if err := mongodb.Update(context.Background(), inviterInfo, update); err != nil {
					log.Error(err)
					return c.Status(fiber.StatusInternalServerError).JSON(presenter.Response{
						Code:    1008,
						Message: err.Error(),
					})
				}
			}
		}

		// 生成邀请分享链接
		payload := fmt.Sprintf("code_%s", account) //邀请人用户ID
		//targetURL := "https://t.me/" + utils.BotUsername + "?start=" + payload    //生成并返回 Telegram 深度链接
		targetURL := "https://t.me/" + utils.BotUsername + "?startapp=" + payload //直接启动小程序
		links := map[string]interface{}{
			"telegramLink": targetURL,
			"facebook":     "https://www.facebook.com/sharer/sharer.php?u=" + url.QueryEscape(targetURL),
			"whatsapp":     "https://wa.me/?text=" + url.QueryEscape(targetURL),
			"telegram":     "https://t.me/share/url?url=" + url.QueryEscape(targetURL),
			"twitter":      "https://twitter.com/intent/tweet?" + "url=" + url.QueryEscape(targetURL),
			"email":        "mailto:?subject=" + url.QueryEscape("Squid Game") + "&body=" + url.QueryEscape(targetURL),
		}
		log.Infof("登录成功, Account:%v, Token:%v, PlatForm:%v", account, sessionToken, loginReq.Type)

		return c.Status(fiber.StatusOK).JSON(presenter.Response{
			Code:    1,
			Message: "Success",
			Result: map[string]interface{}{
				"wsToken": sessionToken,
				"links":   links,
			},
		})
	}
}

// 改用tginit库: "github.com/telegram-mini-apps/init-data-golang"
func Validate(initData, token string, expIn time.Duration) error {
	// 解码查询字符串
	decoded, err := url.QueryUnescape(initData)
	if err != nil {
		fmt.Println("Error decoding query:", err)
		return ErrUnexpectedFormat
	}

	// Parse passed init data as query string.
	params, err := url.ParseQuery(decoded)
	if err != nil {
		return ErrUnexpectedFormat
	}

	var (
		// Init data creation time.
		authDate time.Time
		// Init data sign.
		hash string
		// All found key-value pairs.
		pairs = make([]string, 0, len(params))
	)

	// 指定需要验证的字段
	requiredFields := []string{"user", "chat_instance", "chat_type", "start_param", "auth_date", "hash"}
	for _, field := range requiredFields {
		values, ok := params[field]
		if !ok {
			continue
		}
		value := values[0]
		if field == "hash" {
			hash = value
			continue
		}
		if field == "auth_date" {
			if timestamp, err := strconv.ParseInt(value, 10, 64); err == nil {
				authDate = time.Unix(timestamp, 0)
			}
		}
		pairs = append(pairs, field+"="+value)
	}

	// Sign is always required.
	if hash == "" {
		return ErrSignMissing
	}
	// In case, expiration time is passed, we do additional parameters check.
	// In case, auth date is zero, it means, we can not check if parameters are expired.
	// Check if init data is expired.
	if expIn > 0 && (authDate.IsZero() || authDate.Add(expIn).Before(time.Now())) {
		return ErrExpired
	}
	// According to docs, we sort all the pairs in alphabetical order.
	sort.Strings(pairs)

	joinPairs := strings.Join(pairs, "\n")
	signature := sign(joinPairs, token)

	if signature != hash {
		return ErrSignInvalid
	}
	return nil
}
func sign(payload, key string) string {
	skHmac := hmac.New(sha256.New, []byte("WebAppData"))
	skHmac.Write([]byte(key))
	impHmac := hmac.New(sha256.New, skHmac.Sum(nil))
	impHmac.Write([]byte(payload))
	return hex.EncodeToString(impHmac.Sum(nil))
}
