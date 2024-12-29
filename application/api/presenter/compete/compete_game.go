package compete

type Game struct {
	Id                int32           `bson:"_id"`
	State             int32           `bson:"state"` // 0:开盘中 1:封盘中 2:结算中
	StartTime         int64           `bson:"startTime"`
	CloseTime         int64           `bson:"closeTime"`
	ResultTime        int64           `bson:"resultTime"`
	EndTime           int64           `bson:"endTime"`
	RoundNum          int64           `bson:"roundNum"`
	Amount            int64           `bson:"amount"`     // 可赔付库存
	TakeAmount        int64           `bson:"takeAmount"` // 庄家抽水
	CurRound          *Round          `bson:"curRound"`
	ResultHistoryList []ResultHistory `bson:"resultHistoryList"`
	AccAmountA        AccAmount       `bson:"accAmountA"`     // a累计额度
	AccAmountB        AccAmount       `bson:"accAmountB"`     // b累计额度
	AccAmountPeace    AccAmount       `bson:"accAmountPeace"` // peace累计额度
}
type AccAmount struct {
	AccBet    int64 `bson:"accBet"`    // 游戏累计投注额
	AccPayout int64 `bson:"accPayout"` // 游戏累计赔付额
}

type ResultHistory struct {
	Hit    string  `bson:"hit"`
	Round  int64   `bson:"round"`
	Volume float64 `bson:"volume"`
}

func (game *Game) TableName() string {
	return "compete_game"
}

func (game *Game) PrimaryId() interface{} {
	return game.Id
}

func (game *Game) Countdown(currentTime int64) int64 {
	var countdown int64
	round := game.CurRound
	if game.State == 0 {
		countdown = round.CloseTime - currentTime
	} else if game.State == 1 {
		countdown = round.ResultTime - currentTime
	} else {
		countdown = round.EndTime - currentTime
	}
	return countdown
}
