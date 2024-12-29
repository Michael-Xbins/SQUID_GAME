package presenter

type PumpDetails struct {
	//千分比
	PumpProfit  int64
	PumpJackpot int64
	PumpDay     int64
	PumpActing  int64
	Acting0     int64
	Acting1     int64
	//抽水和奖池分配,以分为单位
	HouseCut              int64 //庄家抽水/机器人抽水
	JackpotContribution   int64 //jackpot(木头人)
	FirstPassContribution int64 //首通抽水(木头人)
	AgentContribution     int64 //代理抽水
	UpLineContribution    int64 //上线
	UpUpLineContribution  int64 //上上线
	AvailableContribution int64 //可赔付库存(玻璃桥)
}
