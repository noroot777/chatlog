package model

// for pad protocol (argly struct...)
type Pad struct {
	UserName string    `json:"userName"`
	AddMsgs  []*PadMsg `json:"AddMsgs"`
}

type PadMsg struct {
	MsgId        string `json:"msg_id"`
	FromUserName *Str   `json:"from_user_name"`
	ToUserName   *Str   `json:"to_user_name"`
	MsgType      int64  `json:"msg_type"`
	SubType      int64  `json:"sub_type,omitempty"`
	Content      *Str   `json:"content"`
	CreateTime   int64  `json:"create_time"`
	NewMsgId     string `json:"new_msg_id"`
}

type Str struct {
	Str string `json:"str"`
}
