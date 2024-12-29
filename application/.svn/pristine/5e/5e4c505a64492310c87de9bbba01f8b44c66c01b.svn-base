package presenter

type UserInfo struct {
	Account        string             `bson:"_id"`
	Nickname       string             `bson:"nickname"`
	SessionToken   string             `bson:"session_token"`
	Balance        int64              `bson:"balance"`        // 玩家代币余额(分为单位)
	USDT           int64              `bson:"USDT"`           // 玩家usdt余额(美分为单位)
	TurnOver       TurnOver           `bson:"turnOver"`       // 游戏币流水
	UsdtRecharge   UsdtRechargeDetail `bson:"usdtTurnOver"`   // usdt充值
	Agent          Agent              `bson:"agent"`          // 游戏币佣金
	UsdtAgent      UsdtDetail         `bson:"usdt_agent"`     // usdt佣金
	UpLine         string             `bson:"upLine"`         // 上线
	DownLines      []string           `bson:"downLines"`      // 下线
	CompletedTasks []string           `bson:"completedTasks"` // 任务完成列表
	Squid          Squid              `bson:"squid"`          // 鱿鱼游戏
	CompeteLastBet int64              `bson:"competeLastBet"` // 拔河本轮下注额(暂存)
	IsRobot        bool               `bson:"isRobot"`        // 机器人
	Welfare        Welfare            `bson:"welfare"`        // 福利金
	Address        string             `bson:"address"`        // 钱包地址
	Voucher        int64              `bson:"voucher"`        // 兑换券
}
type Squid struct {
	RoundId           int32     `bson:"current_round"`              // 玩家当前所在的轮次
	Track             int32     `bson:"current_track"`              // 玩家当前轮次所选的赛道
	BetPrices         int64     `bson:"bet_prices"`                 // 玩家当前轮次的总下注额(以分为单位)
	BetPricesPerRound []int64   `bson:"total_bet_prices_per_round"` // 每轮的总下注额, 重置时机：处理完jackpot后、退赛、死亡
	CanJackpot        bool      `bson:"canJackpot"`                 // 是否可以领取jackpot
	FirstPass         firstPass `bson:"first_pass"`                 // 每日首通奖励池(可累计)
}
type firstPass struct {
	LastFirstPassDate string `bson:"last_first_pass_date"` // 最后一次领取首通奖励的日期
	Pool              int64  `bson:"pool"`                 // 首通奖励资金池：从上一次完成7轮到这一次完成7轮的 每一轮下注额5%的累加 作为奖励池全部返还; 如果一直没有完成第7轮就一直累计5%的首通奖励池,直到某一次完成后全部返还
}
type TurnOver struct {
	LastDate string         `bson:"lastDate"` // 最近更新日期
	Squid    turnOverDetail `bson:"squid"`
	Glass    turnOverDetail `bson:"glass"`
	Ladder   turnOverDetail `bson:"ladder"`
	Compete  turnOverDetail `bson:"compete"`
}
type turnOverDetail struct {
	DailyTurnOver int64 `bson:"dailyTurnOver"` // 每日流水
	TotalTurnOver int64 `bson:"totalTurnOver"` // 总流水 (累计)
}
type UsdtRechargeDetail struct {
	LastDate              string          `bson:"lastDate"`              // 最近更新日期
	DailyRecharge         int64           `bson:"dailyRecharge"`         // 每日充值额
	TotalRecharge         int64           `bson:"totalRecharge"`         // 总充值额 (累计)
	DownLineDailyRecharge map[string]bool `bson:"downLineDailyRecharge"` // 每日充值用户
	DownLineTotalRecharge map[string]bool `bson:"downLineTotalRecharge"` // 所有充值用户
}
type Agent struct {
	LastDate string      `bson:"lastDate"` // 最近更新日期
	Squid    AgentDetail `bson:"squid"`
	Glass    AgentDetail `bson:"glass"`
	Ladder   AgentDetail `bson:"ladder"`
	Compete  AgentDetail `bson:"compete"`
}
type AgentDetail struct {
	DailyAgent int64 `bson:"dailyAgent"` // 每日游戏币佣金
	TotalAgent int64 `bson:"totalAgent"` // 总游戏币佣金 (累计)
	Unclaimed  int64 `bson:"unclaimed"`  // 未领取游戏币佣金
	Claimed    int64 `bson:"claimed"`    // 已领取游戏币佣金
}
type UsdtDetail struct {
	LastDate   string `bson:"lastDate"`   // 最近更新日期
	DailyAgent int64  `bson:"dailyAgent"` // 每日USDT佣金
	TotalAgent int64  `bson:"totalAgent"` // 总USDT佣金 (累计)
	Unclaimed  int64  `bson:"unclaimed"`  // 未领取USDT佣金
	Claimed    int64  `bson:"claimed"`    // 已领取USDT佣金
}
type Welfare struct {
	LastDate    string `bson:"lastDate"`    // 最近更新日期
	Times       int32  `bson:"times"`       // 今日剩余可领次数
	TotalAmount int64  `bson:"totalAmount"` // 累计领取金额
}

func (userinfo *UserInfo) TableName() string {
	return "account"
}

func (userinfo *UserInfo) PrimaryId() interface{} {
	return userinfo.Account
}

type DownLineInfo struct {
	ID                             string             `json:"ID"`
	Name                           string             `json:"Name"`
	UsdtRecharge                   UsdtRechargeDetail `bson:"usdtRecharge"`
	UsdtAgent                      UsdtDetail         `bson:"usdtAgent"`
	DownDownLinesUsdtDailyRecharge int64              `json:"downDownLinesUsdtDailyRecharge"` // 仅下线ID 对应下线的今日充值
	DownDownLinesUsdtRecharge      int64              `json:"downDownLinesUsdtRecharge"`      // 仅下线ID 对应下线的总充值
}
