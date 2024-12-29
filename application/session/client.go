package session

import (
	pb "application/pkg/proto/danmu/message"
	"github.com/gofiber/websocket/v2"
)

//var TokenList = map[string]string{}

var idCounter int64

type Client struct {
	Id           int64
	Account      string
	SessionToken string
	conn         *websocket.Conn
	RespMsg      chan *pb.ClientResponse
	//ReqMsg  chan *pb.ClientRequest
	//Agent *Agent
}

func (client *Client) Conn() *websocket.Conn {
	return client.conn
}

func NewClient(sessionToken, account string, conn *websocket.Conn) *Client {
	idCounter++
	return &Client{
		SessionToken: sessionToken,
		Account:      account,
		conn:         conn,
		RespMsg:      make(chan *pb.ClientResponse, 256),
		Id:           idCounter,
		//ReqMsg:  make(chan *pb.ClientRequest, 1),
		//Agent: &Agent{Id: id},
	}
}
