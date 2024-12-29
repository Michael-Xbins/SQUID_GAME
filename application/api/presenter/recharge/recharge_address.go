package rechargedb

type AddressInfo struct {
	Address    string `bson:"_id"`     // 钱包地址
	UserId     string `bson:"user_id"` // 绑定的的用户ID
	PrivateKey string `bson:"private_key"`
	CreatedAt  int64  `bson:"created_at"`
}

func (info *AddressInfo) TableName() string {
	return "recharge_address"
}

func (info *AddressInfo) PrimaryId() interface{} {
	return info.Address
}
