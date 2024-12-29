package squid

const TotalRounds = 7
const TotalTracks = 4

type GlobalRound struct {
	RoundID int   `bson:"_id"`    // 轮次ID,共有7轮
	Track1  Track `bson:"track1"` // 赛道1
	Track2  Track `bson:"track2"` // 赛道2
	Track3  Track `bson:"track3"` // 赛道3
	Track4  Track `bson:"track4"` // 赛道4
}
type Track struct {
	TotalBetPrices      int64 `bson:"total_bet_prices"`       // 该赛道的总注额
	PlayerNums          int32 `bson:"player_nums"`            // 参与该赛道的玩家数量
	RobotTotalBetPrices int64 `bson:"robot_total_bet_prices"` // 机器人该赛道的总注额
}

func (globalRound *GlobalRound) TableName() string {
	return "squid_round"
}

func (globalRound *GlobalRound) PrimaryId() interface{} {
	return globalRound.RoundID
}
