package glass

type Order struct {
	OrderId   string  `bson:"_id"`
	Account   string  `bson:"account"`
	RoundNum  int64   `bson:"roundNum"`  //期数
	IsWin     bool    `bson:"isWin"`     //是否中奖
	Bonus     int64   `bson:"bonus"`     //奖金
	BetType   string  `bson:"betType"`   //筹码类型
	Odds      float64 `bson:"odds"`      //赔率
	TrackType int     `bson:"trackType"` //一次下注了几个赛道(1|2|3)
	Track1    string  `bson:"track1"`    //赛道1下注的数字
	Track2    string  `bson:"track2"`
	Track3    string  `bson:"track3"`
	Timestamp int64   `bson:"timestamp"`
}

func (order *Order) TableName() string {
	return "glass_order"
}

func (order *Order) PrimaryId() interface{} {
	return order.OrderId
}
