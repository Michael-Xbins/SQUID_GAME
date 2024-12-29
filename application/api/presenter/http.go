package presenter

type Response struct {
	Code    int32                  `json:"code"`
	Message string                 `json:"message"`
	Result  map[string]interface{} `json:"result"`
}

type AppLoginReq struct {
	WebAppInitData string `json:"tgWebAppStartParam"` // 包含邀请人ID(可选字段)
	Type           string `json:"type"`               // web, app
}

type InviteReq struct {
	SessionToken string `json:"sessionToken"`
}

type ClaimReq struct {
	SessionToken string `json:"sessionToken"`
	TaskId       string `json:"taskId"` // 任务ID
}

type GlassBet struct {
	SessionToken string  `json:"sessionToken"`
	Type         int32   `json:"type"` //筹码类型(3个赛道只用一个筹码) 1|2|3|4|5
	Tracks       []Track `json:"tracks"`
}
type Track struct {
	Id  int32 `json:"id"`  // 赛道 (1|2|3)
	Num int   `json:"num"` // 本赛道所选数字id (0|1|2|3)
}

type LadderBet struct {
	SessionToken string  `json:"sessionToken"`
	Orders       []Order `json:"orders"` // 可一次下多个订单
}
type Order struct {
	Id    string            `json:"id"` // 所选数字id (round_10 | round_11 | ... | round_2332)
	Infos []LadderOrderInfo `json:"infos"`
}
type LadderOrderInfo struct {
	Type int32 `json:"type"` // 筹码类型 1|2|3|4|5
	Num  int32 `json:"num"`  // 购买数量
}
