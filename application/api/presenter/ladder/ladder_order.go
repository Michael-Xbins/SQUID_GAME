package ladder

import "application/api/presenter"

type Order struct {
	OrderId    string                      `bson:"_id"`
	Account    string                      `bson:"account"`
	RoundNum   int64                       `bson:"roundNum"`   //期数
	BetId      string                      `bson:"betId"`      //下注ID
	Category   int32                       `bson:"category"`   //直注/二串一
	Infos      []presenter.LadderOrderInfo `bson:"infos"`      // 筹码类型和数量列表
	OrderPrice int64                       `bson:"orderPrice"` //订单总额
	IsWin      bool                        `bson:"isWin"`      //是否中奖
	Bonus      int64                       `bson:"bonus"`      //奖金
	Odds       float64                     `bson:"odds"`       //赔率
	Timestamp  int64                       `bson:"timestamp"`
}

func (order *Order) TableName() string {
	return "ladder_order"
}

func (order *Order) PrimaryId() interface{} {
	return order.OrderId
}
