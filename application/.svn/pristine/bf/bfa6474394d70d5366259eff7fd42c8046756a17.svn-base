package squid

type Fund struct {
	Id              int32     `bson:"_id"`
	HouseCut        int64     `bson:"houseCut"`        // 庄家抽水金额
	Jackpot         int64     `bson:"jackpot"`         // jackpot奖池
	RobotPool       int64     `bson:"robotPool"`       // 机器人抽水金额
	PlayerAccAmount AccAmount `bson:"playerAccAmount"` // 玩家累计额度
	RobotAccAmount  AccAmount `bson:"robotAccAmount"`  // 机器人累计额度
}
type AccAmount struct {
	AccBet    int64 `bson:"accBet"`    // 游戏累计投注额
	AccPayout int64 `bson:"accPayout"` // 游戏累计赔付额
}

func (squidFund *Fund) TableName() string {
	return "squid_fund"
}

func (squidFund *Fund) PrimaryId() interface{} {
	return squidFund.Id
}
