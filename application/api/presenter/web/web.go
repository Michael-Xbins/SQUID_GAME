package web

type Label struct {
	Id        int64  `bson:"_id"`
	Adv       string `bson:"adv"`
	Timestamp int64  `bson:"timestamp"`
}

func (label *Label) TableName() string {
	return "adv_label"
}

func (label *Label) PrimaryId() interface{} {
	return label.Id
}
