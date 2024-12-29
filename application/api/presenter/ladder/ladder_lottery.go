package ladder

type Lottery struct {
	RoundNum  int64  `bson:"_id"` //期数
	Hash      string `bson:"hash"`
	Hash1     int32  `bson:"hash1"`
	Hash2     int32  `bson:"hash2"`
	Hash3     int32  `bson:"hash3"`
	Track1    int32  `bson:"Track1"`    //赛道1开奖的数字
	Track2    int32  `bson:"Track2"`    //赛道2开奖的数字
	Track3    int32  `bson:"Track3"`    //赛道3开奖的数字
	Timestamp int64  `bson:"timestamp"` //开奖时间
}

func (lottery *Lottery) TableName() string {
	return "ladder_lottery"
}

func (lottery *Lottery) PrimaryId() interface{} {
	return lottery.RoundNum
}
