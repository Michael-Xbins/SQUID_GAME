package rechargedb

import "go.mongodb.org/mongo-driver/bson/primitive"

const UsdtToSqu = "UsdtToSqu"
const SquToUsdt = "SquToUsdt"

type ExchangeInfo struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Account   string             `bson:"account"`
	Type      string             `bson:"type"`
	Amount    int64              `bson:"amount"`  // 兑换金额(美分单位)
	Coin      int64              `bson:"coin"`    // 游戏币
	Voucher   int64              `bson:"voucher"` // 兑换券
	CreatedAt int64              `bson:"created_at"`
}

func (info *ExchangeInfo) TableName() string {
	return "exchange_order"
}

func (info *ExchangeInfo) PrimaryId() interface{} {
	return info.ID
}
