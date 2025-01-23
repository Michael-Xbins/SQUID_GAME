package auction

import "go.mongodb.org/mongo-driver/bson/primitive"

type BuyOrderList struct {
	Id             string `bson:"_id"`
	Timestamp      int64  `bson:"timestamp"`
	Account        string `bson:"account"`
	TotalCount     int64  `bson:"totalCount"`
	CompletedCount int64  `bson:"completedCount"`
	Price          int64  `bson:"price"`
}

func (buyOrderList *BuyOrderList) TableName() string {
	return "buy_order_list"
}

func (buyOrderList *BuyOrderList) PrimaryId() interface{} {
	return buyOrderList.Id
}

type SellOrderList struct {
	Id             string `bson:"_id"`
	Timestamp      int64  `bson:"timestamp"`
	Account        string `bson:"account"`
	TotalCount     int64  `bson:"totalCount"`
	CompletedCount int64  `bson:"completedCount"`
	Price          int64  `bson:"price"`
}

func (sellOrderList *SellOrderList) TableName() string {
	return "sell_order_list"
}

func (sellOrderList *SellOrderList) PrimaryId() interface{} {
	return sellOrderList.Id
}

type HistoryList struct {
	Id             primitive.ObjectID `bson:"_id,omitempty"`
	OrderId        string             `json:"orderId"`
	Type           string             `bson:"type"`
	Account        string             `bson:"account"`
	Timestamp      int64              `bson:"timestamp"`
	Price          int64              `bson:"price"`
	CompletedCount int64              `bson:"completedCount"`
	Amount         int64              `bson:"amount"`
}

func (historyList *HistoryList) TableName() string {
	return "history_list"
}

func (historyList *HistoryList) PrimaryId() interface{} {
	return historyList.Id
}

type Game struct {
	Id              int32         `bson:"_id"`
	ClosePrice      int64         `bson:"close_price"` // 昨日收盘价
	MinutesOfPrice  int64         `bson:"minutes_of_price"`
	MinutesOfCount  int64         `bson:"minutes_of_count"`
	TradeRecordList []TradeRecord `bson:"trade_record_list"`
}

type TradeRecord struct {
	Timestamp int64 `bson:"timestamp"`
	AvgPrice  int64 `bson:"avgPrice"`
}

type BuyRecord struct {
	Id    string `bson:"id"`
	Count int64  `bson:"count"`
	Price int64  `bson:"price"`
}

type SellRecord struct {
	Id    string `bson:"id"`
	Count int64  `bson:"count"`
	Price int64  `bson:"price"`
}

func (game *Game) TableName() string {
	return "auction_game"
}

func (game *Game) PrimaryId() interface{} {
	return game.Id
}

//type SellHistoryList struct {
//	Id             string `bson:"_id"`
//	Account        string `bson:"account"`
//	Timestamp      int64  `bson:"timestamp"`
//	Price          int64  `bson:"price"`
//	CompletedCount int64  `bson:"completedCount"`
//	Amount         int64  `bson:"amount"`
//}
//
//func (sellHistoryList *SellHistoryList) TableName() string {
//	return "sell_history_list"
//}
//
//func (sellHistoryList *SellHistoryList) PrimaryId() interface{} {
//	return sellHistoryList.Id
//}
