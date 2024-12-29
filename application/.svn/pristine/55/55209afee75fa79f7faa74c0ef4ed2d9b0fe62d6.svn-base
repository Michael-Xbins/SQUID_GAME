package rechargedb

type Fund struct {
	Id              int32 `bson:"_id"`
	UsdtPool        int64 `bson:"usdtPool"`        // 无上线,无上上线 usdt流向库存(用于对账)
	ServiceFee      int64 `bson:"serviceFee"`      // 提现手续费库存(用于对账)
	WithdrawnAmount int64 `bson:"withdrawnAmount"` // 累计已提现金额
	UsdtToSquCoins  int64 `bson:"usdtToSquCoins"`  // 花费usdt兑换游戏币
	SquToUsdtCoins  int64 `bson:"squToUsdtCoins"`  // 花费游戏币兑换usdt
}

func (info *Fund) TableName() string {
	return "recharge_fund"
}

func (info *Fund) PrimaryId() interface{} {
	return info.Id
}
