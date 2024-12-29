package schedule

import (
	"application/api/handlers/compete"
	"application/api/handlers/ladder"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

const secret = "SECfa0c62e61777031ff6d8cf456341571217a950fd08d5900fd705e32d119157ff"
const webhook = "https://oapi.dingtalk.com/robot/send?access_token=74197e9a9fa3d7909ddaca5eba2c9483378b3eb4be15918e68ed8c1e5c3dc809"

func StartSchedule() {
	utils.LoopSafeGo(scheduleLoop)
}

func scheduleLoop() {
	ticker := time.NewTicker(2 * 60 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			SendToDingTalk()
		}
	}
}

func SendToDingTalk() {
	timestamp := time.Now().UnixNano() / 1e6
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	// 使用 HMAC SHA256 算法生成签名
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(hmac256.Sum(nil))
	// 构建请求 URL，包括 timestamp 和 sign 参数
	webhookURL := fmt.Sprintf("%s&timestamp=%d&sign=%s",
		webhook, timestamp, signature)
	var message string

	// 木头人统计数据
	latest, _, err := redis.GetSquidData()
	if err != nil {
		log.Error("Failed to retrieve latest Squid data:", err)
	}
	if latest != nil {
		env := viper.GetString("common.env")
		if env == utils.Develop {
			env = "内网本地服"
		} else if env == utils.Test {
			env = "外网测试服"
		} else if env == utils.Produce {
			env = "正式服"
		} else {
			env = "服务配置错误"
		}

		message = fmt.Sprintf(
			"  %s  \n"+
				"----------------------------------\n"+
				"📊 **鱿鱼人数据更新** 📊\n"+
				"🕒   %s\n"+
				"💵 【BTC 价格】 %.2f\n"+
				"🔢 【BTC 加和】 %d\n"+
				"🚫 【死亡赛道】 %d\n"+
				"🎲 【赔率1】 %.6f  %d %d\n"+
				"🎲 【赔率2】 %.6f  %d %d\n"+
				"🎲 【赔率3】 %.6f  %d %d\n"+
				"🎲 【赔率4】 %.6f  %d %d\n"+
				"🎲 【赔率5】 %.6f  %d %d\n"+
				"🎲 【赔率6】 %.6f  %d %d\n"+
				"🎲 【赔率7】 %.6f  %d %d\n"+
				"🏆 【jackpot】 %d\n"+
				"💸 【庄家抽成】 %d\n"+
				"💸 【机器人抽成】 %d\n"+
				"🤖 【玩家累计投注额】 %d\n"+
				"💰 【玩家累计赔付额】 %d\n"+
				"🔁 【玩家反水比例】 %.2f\n"+
				"🤖 【机器人累计投注额】 %d\n"+
				"💰 【机器人累计赔付额】 %d\n"+
				"🔁 【机器人反水比例】 %.2f\n"+
				"----------------------------------\n",
			env, latest.Timestamp, latest.BtcPrice, latest.BtcSum, latest.DeadTrack, latest.Odds1, latest.TotalBet1, latest.DeadBet1, latest.Odds2, latest.TotalBet2, latest.DeadBet2,
			latest.Odds3, latest.TotalBet3, latest.DeadBet3, latest.Odds4, latest.TotalBet4, latest.DeadBet4, latest.Odds5, latest.TotalBet5, latest.DeadBet5, latest.Odds6, latest.TotalBet6, latest.DeadBet6,
			latest.Odds7, latest.TotalBet7, latest.DeadBet7, latest.Jackpot, latest.HouseCut, latest.RobotPool, latest.PlayerAccBet, latest.PlayerAccPayout, latest.PlayerRate, latest.RobotAccBet, latest.RobotAccPayout, latest.RobotRate,
		)
	}

	// 梯子统计数据
	if accAmount1, accAmount2, availableFund, err := ladder.GetData(); err == nil {
		sum := accAmount1.AccBet + accAmount2.AccBet
		sumPayout := accAmount1.AccPayout + accAmount2.AccPayout
		backWater := float64(0)
		if sum > 0 {
			backWater = float64(sumPayout) / float64(sum)
		}
		message += fmt.Sprintf(
			"📊 **梯子数据更新** 📊\n"+
				"💰 【直注累计投注总额】 %d\n"+
				"💸 【直注累计赔付总额】 %d\n"+
				"🔄 【直注反水比例】 %.2f\n"+
				"💰 【二串一累计投注总额】 %d\n"+
				"💸 【二串一累计赔付总额】 %d\n"+
				"🔄 【二串一反水比例】 %.2f\n"+
				"💰 【汇总投注总额】 %d\n"+
				"💸 【汇总赔付总额】 %d\n"+
				"🔄 【汇总反水比例】 %.2f\n"+
				"📦 【可赔付库存】 %d\n"+
				"----------------------------------\n",
			accAmount1.AccBet, accAmount1.AccPayout, accAmount1.Rate, accAmount2.AccBet, accAmount2.AccPayout, accAmount2.Rate, sum, sumPayout, backWater, availableFund,
		)
	} else {
		log.Error("ladder.GetData error: ", err)
	}

	// 拔河统计数据
	if accAmountA, accAmountB, accAmountPeace, availableFund, err := compete.GetData(); err == nil {
		sum := accAmountA.AccBet + accAmountB.AccBet + accAmountPeace.AccBet
		sumPayout := accAmountA.AccPayout + accAmountB.AccPayout + accAmountPeace.AccPayout
		backWater := float64(0)
		if sum > 0 {
			backWater = float64(sumPayout) / float64(sum)
		}
		message += fmt.Sprintf(
			"📊 **拔河数据更新** 📊\n"+
				"💰 【A累计投注总额】 %d\n"+
				"💸 【A累计赔付总额】 %d\n"+
				"🔄 【A反水比例】 %.2f\n"+
				"💰 【B累计投注总额】 %d\n"+
				"💸 【B累计赔付总额】 %d\n"+
				"🔄 【B反水比例】 %.2f\n"+
				"💰 【Peace累计投注总额】 %d\n"+
				"💸 【Peace累计赔付总额】 %d\n"+
				"🔄 【Peace反水比例】 %.2f\n"+
				"💰 【汇总投注总额】 %d\n"+
				"💸 【汇总赔付总额】 %d\n"+
				"🔄 【汇总反水比例】 %.2f\n"+
				"📦 【可赔付库存】 %d\n"+
				"----------------------------------\n",
			accAmountA.AccBet, accAmountA.AccPayout, accAmountA.Rate, accAmountB.AccBet, accAmountB.AccPayout, accAmountB.Rate, accAmountPeace.AccBet, accAmountPeace.AccPayout, accAmountPeace.Rate,
			sum, sumPayout, backWater, availableFund,
		)
	} else {
		log.Error("compete.GetData error: ", err)
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	response, err := http.Post(webhookURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Error("Failed to send message to DingTalk:", err)
	}
	defer response.Body.Close()
}
