package session

import (
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils/log"
	"sync/atomic"
)

var counter = int32(0)

type hub struct {
	clientMgr  map[int64]*Client
	register   chan *Client
	unregister chan *Client
	// 函数通道：将函数作为值传递, 这些函数通常会修改共享资源或执行需要同步的操作;
	// 通过将函数发送到这个通道, 可以在一个单独的协程（通常是一个控制协程）中串行地执行这些函数, 从而避免了并发访问和修改共享资源时可能出现的竞态条件;
	// 用于执行任意函数的通道：确保对共享资源（clientMgr 映射）的访问和修改是线程安全的, 即使这些操作来自不同协程
	fChan chan func()
}

var inst = &hub{
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clientMgr:  make(map[int64]*Client),
	fChan:      make(chan func()),
}

func Register(client *Client) {
	inst.register <- client
	v := atomic.AddInt32(&counter, 1)
	S2AllMessage(&pb.ClientResponse{
		Type:    pb.MessageType_OnlineNumNotify_,
		Message: &pb.ClientResponse_OnlineNumNotify{OnlineNumNotify: &pb.OnlineNumNotify{Count: v}},
	})
}

func Unregister(client *Client) {
	inst.unregister <- client
	v := atomic.AddInt32(&counter, -1)
	S2AllMessage(&pb.ClientResponse{
		Type:    pb.MessageType_OnlineNumNotify_,
		Message: &pb.ClientResponse_OnlineNumNotify{OnlineNumNotify: &pb.OnlineNumNotify{Count: v}},
	})
}

var S2AllMessage = func(response *pb.ClientResponse) {
	inst.fChan <- func() {
		for _, agent := range inst.clientMgr {
			select {
			case agent.RespMsg <- response:
			default:
				log.Error("网络数据发送阻塞", response.Type)
			}
		}
	}
}

var S2CMessage = func(account string, response *pb.ClientResponse) {
	inst.fChan <- func() {
		for _, agent := range inst.clientMgr {
			if agent.Account == account {
				select {
				case agent.RespMsg <- response:
				default:
					log.Error("网络数据发送阻塞", response.Type)
				}
			}
		}
	}
}

func Run() {
	h := inst
	for {
		select {
		case client := <-h.register:
			h.clientMgr[client.Id] = client
		case client := <-h.unregister:
			if _, ok := h.clientMgr[client.Id]; ok {
				delete(h.clientMgr, client.Id)
				close(client.RespMsg)
			}
		case f := <-h.fChan:
			f() // 执行fChan通道接收到的函数
		}
	}
}
