package main

import (
	"application/api/handlers"
	"application/api/handlers/compete"
	"application/api/handlers/ladder"
	"application/api/handlers/schedule"
	"application/api/handlers/squid"
	"application/api/routes"
	"application/binance"
	"application/mongodb"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"application/socket"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/logger"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var (
	cancelFunc context.CancelFunc
)

const (
	defaultStackSize = 4096
)

func init() {
	//utils.InitTimeZone()
	log.InitConfig()
	utils.InitLubanTables()
}

// getCurrentGoroutineStack 获取当前Goroutine的调用栈，便于排查panic异常
func getCurrentGoroutineStack() string {
	var buf [defaultStackSize]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}

func setupMiddlewares(app *fiber.App) {
	logDir := "log" // 确保log目录存在
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.Mkdir(logDir, 0755); err != nil {
			log.Fatalf("failed to create log directory: %v", err)
		}
	}
	logFilePath := filepath.Join(logDir, "access.log")
	accessLogFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	app.Use(
		fiberRecover.New(),
		requestid.New(),
		logger.New(logger.Config{
			Output:     accessLogFile,
			Format:     "[${time}] ${ip}:${port} ${status}${method}${header:Playerid}| ReqID:${locals:requestid} | Path:${path} | Body:${body} | ${queryParams} ${latency}\n",
			TimeFormat: "2006-01-02 15:04:05",
		}),
		// 添加全局CORS中间件
		func(c *fiber.Ctx) error {
			c.Set("Content-Type", "application/json")
			c.Set("Access-Control-Allow-Origin", "*")
			c.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			c.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			if c.Method() == "OPTIONS" {
				return c.SendStatus(http.StatusNonAuthoritativeInfo) //用于CORS预检
			}
			return c.Next()
		},
	)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("[panic] err: %v\nstack: %s\n", err, getCurrentGoroutineStack())
		}
	}()
	var err error
	err = redis.NewRedisClient() // 初始化redis
	if err != nil {
		panic(fmt.Errorf("fatal on connect redis: %w", err))
	}
	defer func() {
		redis.CloseRedisClient()
	}()
	cancelFunc, err = mongodb.NewMongoDBClient() // 初始化mongoDB
	if err != nil {
		panic(fmt.Errorf("fatal on connect mongodb: %w", err))
	}
	defer cancelFunc()
	app := fiber.New()
	setupMiddlewares(app)
	setupRoutes(app)
	binance.StartPullBinance()
	squid.StartSquidGame()
	compete.StartCompeteGame()
	//glass.StartGlassGame()
	ladder.StartLadderGame()
	schedule.StartSchedule()
	if err := app.Listen(fmt.Sprintf(":%d", viper.GetInt("common.port"))); err != nil {
		log.Fatal(err)
	}
}

func setupRoutes(app *fiber.App) {
	err := socket.NewWebSocketServer(&socket.WebSocketHandler{},
		socket.WithReadBuffer(2048),
		socket.WithWriteBuffer(2048)).StartWs(app)
	if err != nil {
		log.Panic(err)
	}

	//app
	app.Post("/app/login/http", handlers.LoginHandler())
	routes.AppRouter(app.Group("/app/game"))

	//鱿鱼游戏
	routes.SquidRouter(app.Group("/app/squid/game"))

	//拔河游戏
	routes.CompeteRouter(app.Group("/app/compete/game"))

	//玻璃桥游戏
	//routes.GlassRouter(app.Group("/app/glass/game"))

	//梯子游戏
	routes.LadderRouter(app.Group("/app/ladder/game"))
}
