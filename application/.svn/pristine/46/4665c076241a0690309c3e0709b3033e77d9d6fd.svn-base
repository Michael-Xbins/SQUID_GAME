package compete

type Round struct {
	Number int64             `bson:"_id"`
	Orders map[string]*Order `bson:"orders"` //K:账号 V:下注的信息

	StartTime  int64     `bson:"startTime"`
	CloseTime  int64     `bson:"closeTime"`
	ResultTime int64     `bson:"resultTime"`
	EndTime    int64     `bson:"endTime"`
	Nums       []float64 `bson:"nums"`
	Sum        float64   `bson:"sum"`
	//EndNum     float64   `bson:"endNum"`
}

type Order struct {
	Amounts map[string]int64 `bson:"amounts"` //K:a,b,peace V:值
}

func (round *Round) TableName() string {
	return "round"
}

func (round *Round) PrimaryId() interface{} {
	return round.Number
}
