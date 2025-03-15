package data

type XYMsg struct {
	Ack  int32  `json:"ack"`
	Body any    `json:"body"`
	Cmd  string `json:"cmd"`
	Seq  int32  `json:"seq"`
	Time int64  `json:"time"`
}
