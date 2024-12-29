package socket

import (
	"application/mongodb"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils/log"
	"application/redis"
	"application/session"
	"context"
	"fmt"
	"github.com/gofiber/websocket/v2"
	"github.com/golang/protobuf/proto"
	"sync"
	"time"
)

const (
	writeWait = 240 * time.Second // 写操作的超时时间

	pongWait = 180 * time.Second // 等待客户端pong响应的最大时间

	pingPeriod     = (pongWait * 9) / 10 // 发送Ping消息的周期
	maxMessageSize = 1024 * 16           // WebSocket消息的最大容量
)

type WebSocketHandler struct {
}

var activeConnections = make(map[string]*websocket.Conn) // 全局变量,存储用户账号和其WebSocket的映射
var connMutex sync.Mutex

func (handler *WebSocketHandler) HandleConn(token string, conn *websocket.Conn) {
	userInfo, err := mongodb.FindUserInfoByToken(context.Background(), token)
	if err != nil {
		// 发送Token错误消息
		errMsg := fmt.Sprintf("Invalid token, 请重新登录: %s", err.Error())
		if err := conn.WriteMessage(websocket.TextMessage, []byte(errMsg)); err != nil {
			log.Error("Error sending error message:", err)
		}
		//关闭websocket连接
		conn.Close()
		log.Errorf("token:%s, user:%s, HandleConn error:%s", token, userInfo.Account, err)
		return
	}

	// 多设备同时登陆, 删除旧设备连接
	connMutex.Lock()
	if oldConn, exists := activeConnections[userInfo.Account]; exists {
		session.S2CMessage(userInfo.Account, &pb.ClientResponse{
			Type:    pb.MessageType_CloseConnNotify_,
			Message: &pb.ClientResponse_CloseConnNotify{CloseConnNotify: &pb.CloseConnNotify{}},
		})
		oldConn.Close()
		log.Infof("多设备同时登陆账户:%s, 删除旧设备链接:%s, 新链接:%s", userInfo.Account, oldConn.RemoteAddr(), conn.RemoteAddr())
	}
	activeConnections[userInfo.Account] = conn
	connMutex.Unlock()

	agent := session.NewClient(userInfo.SessionToken, userInfo.Account, conn)
	log.Infof("[%s]连接成功\n", conn.RemoteAddr().String())
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		writePump(agent)
	}()
	go func() {
		defer wg.Done()
		readPump(agent)
	}()
	wg.Wait()    // 等待所有协程完成
	conn.Close() // 所有协程完成后关闭连接
}

func (handler *WebSocketHandler) HandleStarted() {
	go session.Run()
}
func (handler *WebSocketHandler) HandleStopped() {
}

func writePump(client *session.Client) {
	ticker := time.NewTicker(pingPeriod) // 定期发送WebSocket的Ping消息,周期为pingPeriod
	sessionTickTime := time.Minute * 30
	sessionTicker := time.NewTicker(sessionTickTime * 2 / 3)
	redis.SetSessionToken(client.SessionToken, client.Account, sessionTickTime) // 定时更新Redis中的会话令牌,周期为30分钟
	defer func() {
		ticker.Stop()
		session.S2CMessage(client.Account, &pb.ClientResponse{
			Type:    pb.MessageType_CloseConnNotify_,
			Message: &pb.ClientResponse_CloseConnNotify{CloseConnNotify: &pb.CloseConnNotify{}},
		})
		// 清理连接映射
		connMutex.Lock()
		delete(activeConnections, client.Account)
		connMutex.Unlock()

		client.Conn().Close()
		redis.DelSessionToken(client.SessionToken)
		log.Infof("[%s]连接断开\n", client.Conn().RemoteAddr().String())
	}()
	for {
		select {
		case message, ok := <-client.RespMsg:
			client.Conn().SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.Conn().WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			data, err := proto.Marshal(message)
			if err != nil {
				log.Errorf("消息解析错误:[%v]", message)
				return
			}
			err = client.Conn().WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				return
			}
		case <-ticker.C:
			client.Conn().SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn().WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-sessionTicker.C:
			redis.SetSessionToken(client.SessionToken, client.Account, sessionTickTime*2/3)
		}
	}
}

func readPump(client *session.Client) {
	session.Register(client)
	defer func() {
		log.Infof("[%s]连接断开", client.Conn().RemoteAddr().String())
		session.Unregister(client)
		client.Conn().Close()
	}()
	client.Conn().SetReadLimit(maxMessageSize)
	client.Conn().SetReadDeadline(time.Now().Add(pongWait)) // 客户端必须在指定的 pongWait 时间内发送数据到svr,否则svr将无法从该连接读取任何数据
	for {
		_, data, err := client.Conn().ReadMessage() // 若超时pongWait,返回超时错误
		if err != nil {
			log.Errorf("error: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error: %v", err)
			}
			break
		}
		client.Conn().SetReadDeadline(time.Now().Add(pongWait)) // 绝对时间：读取完需要重新设置
		msg := &pb.ClientRequest{}
		if err := proto.Unmarshal(data, msg); err != nil {
			log.Error("消息解析异常", err)
			return
		}
		if msg.GetType() == pb.MessageType_HeartBeatRequest_ {
			session.S2CMessage(client.Account, &pb.ClientResponse{
				Type:    pb.MessageType_HeartBeatResponse_,
				Message: &pb.ClientResponse_HeartBeatResponse{HeartBeatResponse: &pb.HeartBeatResponse{}},
			})
		}
	}
}
