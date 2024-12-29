package compete

type OrderInfo struct {
	OrderId     string  `bson:"_id"`
	Account     string  `bson:"account"`
	RoundNum    int64   `bson:"roundNum"`    //期数
	TransAmount float64 `bson:"transAmount"` //BTC交易量
	Hit         string  `bson:"hit"`         //胜利赛道
	Track       string  `bson:"track"`       //所选赛道
	IsWin       bool    `bson:"isWin"`       //是否中奖
	BetAmount   int64   `bson:"betAmount"`   //下注额
	Odds        int64   `bson:"odds"`        //赔率
	Bonus       int64   `bson:"bonus"`       //奖金
	Timestamp   int64   `bson:"timestamp"`
}

func (orderInfo *OrderInfo) TableName() string {
	return "compete_order"
}

func (orderInfo *OrderInfo) PrimaryId() interface{} {
	return orderInfo.OrderId
}
