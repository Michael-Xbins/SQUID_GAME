package squid

// Game 游戏相关配置,倒计时状态
type Game struct {
	Id            int32   `bson:"_id"`
	RoundNum      int64   `bson:"roundNum"`
	State         int32   `bson:"state"` // 0:开盘中 1:封盘中 2:结算中
	StartTime     int64   `bson:"startTime"`
	CloseTime     int64   `bson:"closeTime"`
	DeadTrackTime int64   `bson:"deadTrackTime"`
	ResultTime    int64   `bson:"resultTime"`
	EndTime       int64   `bson:"endTime"`
	TransPrice    float64 `bson:"transPrice"`
	DeadTrack     int32   `bson:"deadTrack"`
}

func (game *Game) TableName() string {
	return "squid_game"
}

func (game *Game) PrimaryId() interface{} {
	return game.Id
}

func (game *Game) SquidCountdown(currentTime int64) int64 {
	var countdown int64
	if game.State == 0 {
		countdown = game.CloseTime - currentTime
	} else if game.State == 1 {
		countdown = game.ResultTime - currentTime
	} else {
		countdown = game.EndTime - currentTime
	}
	return countdown
}
