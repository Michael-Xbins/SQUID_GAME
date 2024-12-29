package glass

type Fund struct {
	Id            int32     `bson:"_id"`
	HouseCut      int64     `bson:"houseCut"`       // 庄家抽水金额
	AvailableFund int64     `bson:"available_fund"` // 可赔付库存
	AccAmount1    AccAmount `bson:"accAmount1"`     // 直注累计额度
	AccAmount2    AccAmount `bson:"accAmount2"`     // 二串一累计额度
	AccAmount3    AccAmount `bson:"accAmount3"`     // 三串一累计额度
}
type AccAmount struct {
	AccBet    int64 `bson:"accBet"`    // 游戏累计投注额
	AccPayout int64 `bson:"accPayout"` // 游戏累计赔付额
}

func (glassFund *Fund) TableName() string {
	return "glass_fund"
}

func (glassFund *Fund) PrimaryId() interface{} {
	return glassFund.Id
}
