package presenter

type CdkInfo struct {
	Cdk            string `bson:"_id"`
	Salt           string `bson:"salt"`
	ExchangeID     int    `bson:"exchangeID"`
	ExchangeAmount int64  `bson:"exchangeAmount"`
	Received       bool   `bson:"Received"`
}

func (cdkInfo *CdkInfo) TableName() string {
	return "cdk"
}

func (cdkInfo *CdkInfo) PrimaryId() interface{} {
	return cdkInfo.Cdk
}
