package socket

import (
	"github.com/gofiber/fiber/v2"
	//"framework/component"
	"github.com/gofiber/websocket/v2"
)

type Server struct {
	cfg     *Config
	handler IWebSocketServerHandler
}

type IWebSocketServerHandler interface {
	HandleConn(token string, conn *websocket.Conn)
	HandleStarted()
	HandleStopped()
}

type Option func(cfg *Config)

type Config struct {
	addr        string
	network     string
	readBuffer  int
	writeBuffer int
	serviceName string
}

func WithReadBuffer(readBuffer int) Option {
	return func(cfg *Config) {
		cfg.readBuffer = readBuffer
	}
}

func WithWriteBuffer(writeBuffer int) Option {
	return func(cfg *Config) {
		cfg.writeBuffer = writeBuffer
	}
}

func WithAddress(addr string) Option {
	return func(cfg *Config) {
		cfg.addr = addr
	}
}
func WithNetwork(network string) Option {
	return func(cfg *Config) {
		cfg.network = network
	}
}
func WithServiceName(serviceName string) Option {
	return func(cfg *Config) {
		cfg.serviceName = serviceName
	}
}

func NewWebSocketServer(handler IWebSocketServerHandler, opts ...Option) *Server {
	cfg := &Config{
		//addr:        "localhost:8080",
		//serviceName: "server",
		network:     "tcp",
		readBuffer:  1024,
		writeBuffer: 4096,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Server{
		cfg:     cfg,
		handler: handler,
	}
}

func (ws *Server) StartWs(app *fiber.App) error {
	//中间件, 检查传入的请求是否 WebSocket 升级请求
	app.Use("/app/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	//建立websocket连接, websocket.New 封装了 upgrader.Upgrade(): 升级HTTP到WebSocket
	app.Get("/app/ws", websocket.New(func(conn *websocket.Conn) {
		if ws.handler != nil {
			token := conn.Query("token")
			if token == "" {
				conn.Close()
				return
			}
			ws.handler.HandleConn(token, conn) //处理websocket连接
		}
	}))
	if ws.handler != nil {
		ws.handler.HandleStarted()
	}
	return nil
}

func (ws *Server) StopWs() error {
	if ws.handler != nil {
		ws.handler.HandleStopped()
	}
	return nil
}

//func (ws *Server) Start() error {
//	ws.upgrader = websocket.Upgrader{
//		ReadBufferSize:  ws.cfg.readBuffer,
//		WriteBufferSize: ws.cfg.writeBuffer,
//		CheckOrigin: func(r *http.Request) bool {
//			return true
//		},
//	}
//	// 添加WebSocket路由处理程序
//	http.HandleFunc("/app/ws", func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Access-Control-Allow-Origin", "*")
//		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
//		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
//		ws.serveWs(w, r)
//	})
//	if ws.handler != nil {
//		ws.handler.HandleStarted()
//	}
//	return nil
//}
//
//func (ws *Server) Stop() error {
//	if ws.handler != nil {
//		ws.handler.HandleStopped()
//	}
//	return nil
//}
//
//// serveWs handles websocket requests from the peer.
//func (ws *Server) serveWs(w http.ResponseWriter, r *http.Request) {
//	conn, err := ws.upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		log.Println(err)
//		return
//		//ws://xxx.x.xxx.xx:000/ws?token=xxxxx
//	}
//	if ws.handler != nil {
//		token := r.URL.Query().Get("token")
//		if token == "" {
//			conn.Close()
//			return
//		}
//		ws.handler.HandleConn(token, conn)
//	}
//}
