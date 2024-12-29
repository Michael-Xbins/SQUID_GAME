package squid

type Order struct {
	OrderId    string  `bson:"_id"`
	IsRobot    bool    `bson:"isRobot"` // 机器人
	Account    string  `bson:"account"`
	RoundNum   int64   `bson:"roundNum"`   //期数
	Track      int32   `bson:"track"`      //所选赛道
	BetPrices  int64   `bson:"betPrices"`  //下注额
	TransPrice float64 `bson:"transPrice"` //BTC价格
	BtcSum     int32   `bson:"btc_sum"`    //BTC价之和
	Odds       float64 `bson:"odds"`       //赔率
	IsWin      bool    `bson:"isWin"`      //是否中奖
	Bonus      int64   `bson:"bonus"`      //奖金
	Timestamp  int64   `bson:"timestamp"`
}

func (order *Order) TableName() string {
	return "squid_order"
}

func (order *Order) PrimaryId() interface{} {
	return order.OrderId
}
