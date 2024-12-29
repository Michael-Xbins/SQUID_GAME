package rechargedb

type PayInfo struct {
	Hash           string `bson:"_id"` //交易哈希
	Account        string `bson:"account"`
	Network        string `bson:"network"`
	Address        string `bson:"address"`        // 代币合约地址
	Value          string `bson:"value"`          // 交易金额
	Symbol         string `bson:"symbol"`         // 代币符号
	Decimals       int    `bson:"decimals"`       // 代币小数位
	Name           string `bson:"name"`           // 代币名称
	GameCoins      int64  `bson:"gameCoins"`      // 兑换的游戏币
	Voucher        int64  `bson:"voucher"`        // 兑换的券
	BlockTimestamp int64  `bson:"blockTimestamp"` // 到账时间d
	CreatedAt      int64  `bson:"created_at"`
}

func (info *PayInfo) TableName() string {
	return "recharge_order"
}

func (info *PayInfo) PrimaryId() interface{} {
	return info.Hash
}
