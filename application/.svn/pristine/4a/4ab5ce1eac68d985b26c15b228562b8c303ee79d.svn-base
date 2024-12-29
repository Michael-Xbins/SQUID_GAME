package routes

import (
	"application/api/handlers"
	"application/api/handlers/compete"
	"application/api/handlers/glass"
	"application/api/handlers/ladder"
	"application/api/handlers/squid"
	"github.com/gofiber/fiber/v2"
)

func AppRouter(app fiber.Router) {
	app.Post("/invited/http", handlers.GetInvitesHandler()) // 获取邀请次数
	app.Post("/claim_invite_reward/http", handlers.ClaimInviteReward())
	app.Post("/userinfo/http", handlers.GetUserInfo())
	app.Post("/set_name/http", handlers.SetName())
	app.Post("/claim_agent/http", handlers.ClaimAgent()) // 领取代理usdt
	//app.Post("/get_cdk/http", handlers.GetCDK())
	app.Post("/exchange_cdk/http", handlers.ExchangeCDK())          // 兑换CDK
	app.Post("/service/http", handlers.MyService())                 // 我的客服
	app.Post("/claim_welfare_reward/http", handlers.ClaimWelfare()) // 领取福利金奖励
	app.Post("/get_welfare_count/http", handlers.GetWelfare())
	app.Post("/get_asset/http", handlers.GetAsset())

	app.Post("/recharge/http", handlers.Recharge())            // 充值
	app.Post("/withdrawal/http", handlers.Withdrawal())        // 提现
	app.Post("/exchange_usdt_to_squ/http", squid.UsdtToSqu())  // USDT 兑换 兑换券和游戏币 (1:10:1000000)
	app.Post("/exchange_squ_to_usdt/http", squid.SquToUsdt())  // 兑换券和游戏币 兑换 USDT
	app.Post("/history_record/http", handlers.HistoryRecord()) // 历史记录
}

func SquidRouter(app fiber.Router) {
	app.Post("/status/http", squid.Status())                // 游戏状态
	app.Post("/order/http", squid.Order())                  // 下注
	app.Post("/cancel/http", squid.Cancel())                // 取消下注
	app.Post("/switch/http", squid.Switch())                // 转换赛道
	app.Post("/history_orders/http", squid.HistoryOrders()) // 历史订单
	app.Post("/data/http", squid.Data())
	app.Post("/first_pass_status/http", squid.FirstPassStatus())
}

func CompeteRouter(app fiber.Router) {
	app.Post("/order/http", compete.Order)
	app.Post("/cancel/http", compete.Cancel)
	app.Post("/history/http", compete.History)
	app.Post("/history_orders/http", compete.HistoryOrders()) // 历史订单
	app.Post("/data/http", compete.Data())
}

func GlassRouter(app fiber.Router) {
	app.Post("/status/http", glass.Status())                // 游戏状态
	app.Post("/order/http", glass.Order())                  // 下注
	app.Post("/history_orders/http", glass.HistoryOrders()) // 历史订单
	app.Post("/lottery/http", glass.Lottery())              // 趋势图
	app.Post("/data/http", glass.Data())
}

func LadderRouter(app fiber.Router) {
	app.Post("/status/http", ladder.Status())                // 游戏状态
	app.Post("/order/http", ladder.Order())                  // 下注
	app.Post("/history_orders/http", ladder.HistoryOrders()) // 历史订单
	app.Post("/lottery/http", ladder.Lottery())              // 趋势图
	app.Post("/data/http", ladder.Data())
}
