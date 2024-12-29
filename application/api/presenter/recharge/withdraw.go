package rechargedb

type WithDrawInfo struct {
	OrderId    string `bson:"_id"`    // 交易id
	TxHash     string `bson:"txHash"` // 交易哈希
	Account    string `bson:"account"`
	Network    string `bson:"network"`
	Address    string `bson:"address"`    // 转账地址
	Amount     int64  `bson:"amount"`     // 转账金额(美分单位)
	ServiceFee int64  `bson:"serviceFee"` // 手续费(美分单位)
	RealAmount int64  `bson:"realAmount"` // 真正转账金额(美分单位)
	CreatedAt  int64  `bson:"created_at"`
}

func (info *WithDrawInfo) TableName() string {
	return "withdraw_order"
}

func (info *WithDrawInfo) PrimaryId() interface{} {
	return info.OrderId
}
