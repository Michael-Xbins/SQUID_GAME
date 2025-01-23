package presenter

type Cdk struct {
	Cdk         string         `bson:"_id"`
	BatchNum    int64          `bson:"batchNum"`    // 批次号
	Type        string         `bson:"type"`        // assign, single, multi
	Deadline    int64          `bson:"deadline"`    // 截止日期
	Channel     string         `bson:"channel"`     // 生效渠道
	IsRecharge  bool           `bson:"isRecharge"`  // 是否为充值后才可用
	Times       int            `bson:"times"`       // 限用次数
	Rewards     []RewardInfo   `bson:"rewards"`     // cdk兑换奖励列表
	Salt        string         `bson:"salt"`        // 盐值
	ExchangeUid map[string]int `bson:"exchangeUid"` // 已兑换名单和已使用次数
}
type RewardInfo struct {
	RewardId  string `bson:"rewardId"`  // 奖励ID (coin, voucher)
	RewardNum int64  `bson:"rewardNum"` // 奖励数量
}

func (cdkInfo *Cdk) TableName() string {
	return "cdk"
}

func (cdkInfo *Cdk) PrimaryId() interface{} {
	return cdkInfo.Cdk
}

type BatchCdk struct {
	Num       int64             `bson:"_id"` //批次号
	CDKs      []string          `bson:"cdks"`
	SingleUid map[string]string `bson:"singleUid"` //单人CDK: (1)其他玩家互斥; (2)与本批次中其他码互斥.
	Timestamp int64             `bson:"timestamp"`
}

func (info *BatchCdk) TableName() string {
	return "cdk_batch"
}

func (info *BatchCdk) PrimaryId() interface{} {
	return info.Num
}
