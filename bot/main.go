package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"strings"
)

const (
	TelegramToken       = "7369353290:AAEgnAXlCIXdHMOhy2AZoqaD5pOB_fj5Hi4" //替换为你的机器人Token
	BaseWebAppURL       = "https://www.giantbee.net/squid/"                //客户端访问地址
	ServerURL           = "https://www.giantbee.net/bot/"
	DetailedDescription = "🎉 Novice rewards with gifts, truly valuable tokens\n\n" +
		"💹 Hello, I have received an invitation letter from 'Squid Game':\nWooden Man, Glass Bridge, Tug of War Game\nWe sincerely invite you to experience the most authentic and interesting life and death game!\nWin huge bonuses by clearing the level!\n\n" +
		"🔮 3000 SQU per day for everyone to play with\nParticipate in the Woodman game to divide and win the jackpot prize\nParticipating in tug of war competitions quickly doubles assets\nParticipate in ladder games with multiple gameplay options and receive up to 10 times the bonus refund\n\n" +
		"🎯 The popular online drama \"Squid Game\" with the same name has been launched!\nFairness, Truth, and Fun\nUsing cryptocurrency as the winning result\nQuick Withdrawal: T+0 to Bank Account\nInviting partners can become advanced agents\nClick to enter and win huge bonuses!\n"
)

// 轮询方式
//func main() {
//	bot, err := tgbotapi.NewBotAPI(TelegramToken) // 请替换为你的机器人Token
//	if err != nil {
//		log.Panic(err)
//	}
//	bot.Debug = true
//	u := tgbotapi.NewUpdate(0)
//	u.Timeout = 60
//
//	updates := bot.GetUpdatesChan(u)
//	for update := range updates {
//		if update.Message != nil { // 确保收到的是消息
//			if update.Message.IsCommand() { // 检查消息是否为命令
//				switch update.Message.Command() {
//				case "start":
//					startCommandHandler(update, bot)
//				case "pin":
//					pinCommandHandler(update, bot)
//				}
//			}
//		}
//	}
//}

// webhook方式
func main() {
	bot, err := tgbotapi.NewBotAPI(TelegramToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	// 取消 Webhook
	params := tgbotapi.Params{
		"url": "",
	}
	_, err = bot.MakeRequest("setWebhook", params)
	if err != nil {
		log.Fatal("Failed to clear webhook: ", err)
	} else {
		log.Println("Webhook successfully removed.")
	}

	// 手动设置 Webhook
	webhookURL := ServerURL + bot.Token
	params = tgbotapi.Params{
		"url": webhookURL,
	}
	_, err = bot.MakeRequest("setWebhook", params)
	if err != nil {
		log.Fatal(err)
	}

	webhookInfo, err := bot.GetWebhookInfo()
	if err != nil {
		log.Println("Failed to get webhook info:", err)
		return
	}
	log.Printf("Webhook Info: %+v\n", webhookInfo)

	updates := bot.ListenForWebhook("/bot/" + bot.Token) //监听 Webhook
	go func() {
		err := http.ListenAndServe(":8083", nil)
		if err != nil {

		}
	}() // 中台nginx反向代理处理https

	// 处理更新
	for update := range updates {
		if update.Message != nil { // 确保收到的是消息
			if update.Message.IsCommand() { // 检查消息是否为命令
				switch update.Message.Command() {
				case "start":
					startCommandHandler(update, bot)
				case "pin":
					pinCommandHandler(update, bot)
				}
			}
		}
	}
}

func startCommandHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if isAdmin(update, bot) && (update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()) { // 确保消息来自群组,并且拥有权限
		// 群组中发送包含 URL 按钮的消息
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, DetailedDescription+"\n"+"Click on the link below to open the game application in the browser：")
		targetURL := "https://t.me/" + "XbinSky_bot"
		btn := tgbotapi.NewInlineKeyboardButtonURL("START GAME！", targetURL)
		markup := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})
		msg.ReplyMarkup = markup
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message with URL button: %v", err)
		}
	}
	// 解析通过深度链接传递的参数
	args := update.Message.CommandArguments()
	fmt.Println("Received args: ", args)
	parts := strings.Split(args, "_")
	var inviterID string

	// 检查参数格式是否正确
	if len(parts) > 1 && parts[0] == "code" {
		inviterID = parts[1]
		log.Printf("Inviter ID: %s", inviterID)
	}

	// 构建 Web 应用的 URL
	webAppURL := BaseWebAppURL
	if inviterID != "" {
		webAppURL += "?code=" + inviterID // 如果有邀请人ID, 则附加到 URL
	}

	if _, err := sendWebAppButton(update, bot, webAppURL); err != nil {
		log.Printf("sendWebAppButton err: %v", err)
	}
}

func pinCommandHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if isAdmin(update, bot) && (update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()) { // 确保消息来自群组,并且拥有权限
		// 群组中发送包含 URL 按钮的消息
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, DetailedDescription+"\n"+"Click on the link below to open the game application in the browser：")
		targetURL := "https://t.me/" + "XbinSky_bot"
		btn := tgbotapi.NewInlineKeyboardButtonURL("START GAME！", targetURL)
		markup := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})
		msg.ReplyMarkup = markup
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message with URL button: %v", err)
		}
	} else {
		webAppURL := BaseWebAppURL
		sendMsg, err := sendWebAppButton(update, bot, webAppURL)
		if err != nil {
			log.Printf("sendWebAppButton err: %v", err)
			return
		}
		// 固定消息
		pinConfig := tgbotapi.PinChatMessageConfig{
			ChatID:    update.Message.Chat.ID,
			MessageID: sendMsg.MessageID,
		}
		if _, err := bot.Request(pinConfig); err != nil {
			log.Printf("Error pinning message: %v", err)
		}
	}
}

func sendWebAppButton(update tgbotapi.Update, bot *tgbotapi.BotAPI, webAppURL string) (tgbotapi.Message, error) {
	// 创建一个按钮，用户点击后可以打开 Web 应用
	webAppButton := tgbotapi.InlineKeyboardButton{
		Text:   "PLAY!",
		WebApp: &tgbotapi.WebAppInfo{URL: webAppURL}, // 使用 WebApp 属性指向你的 Web App URL
	}
	markup := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{webAppButton})

	// 创建并发送 带有按钮的消息至 Telegram 客户端
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, DetailedDescription)
	msg.ReplyMarkup = markup
	sentMsg, err := bot.Send(msg)
	if err != nil {
		return sentMsg, err
	}
	return sentMsg, nil
}
func isAdmin(update tgbotapi.Update, bot *tgbotapi.BotAPI) bool {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	chatMemberConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	}
	chatMember, err := bot.GetChatMember(chatMemberConfig)
	if err != nil {
		log.Printf("Error getting chat member: %v", err)
		return false
	}
	return chatMember.IsAdministrator() || chatMember.IsCreator()
}
