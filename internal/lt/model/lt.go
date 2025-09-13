package model

// for members update
type Members struct {
	Tz     string            `json:"tz"`
	Groups []*MembersInGroup `json:"groups"`
}

type MembersInGroup struct {
	Wxid     string    `json:"wxid"`
	Roomname string    `json:"roomname"`
	Members  []*Member `json:"members"`
}

type Member struct {
	Wxid     string `json:"wxid"`
	Nickname string `json:"nickname"`
	// Displayname string `json:"displayname"`
	// Sendname    string `json:"sendname"`
	Remark       string `json:"remark"`
	Describe     string `json:"describe"`
	Lalels       string `json:"labels"`
	Roomnickname string `json:"roomnickname,omitempty"`
}

// for group msg
type GM struct {
	Tz   string `json:"tz"`
	Room string `json:"room"`
	Wxid string `json:"wxid"`
	Msgs []*Msg `json:"msgs"`
}

type Msg struct {
	Wxid     string `json:"wxid"`
	Nickname string `json:"nickname"`
	// Remark       string `json:"remark"`
	// Describe     string `json:"describe,omitempty"`
	// Labels       string `json:"labels,omitempty"`
	// Roomnickname string `json:"roomnickname,omitempty"`
	When string `json:"when"`
	Msg  string `json:"msg"`
}

// for cash msg
type Cashes struct {
	Tz       string  `json:"tz"`
	Wxid     string  `json:"wxid"`
	CashList []*Cash `json:"cash_list"`
}

type Cash struct {
	Who           string `json:"who"`
	Type          string `json:"type"`
	When          string `json:"when"`
	Amount        string `json:"amount"`
	Memo          string `json:"memo"`
	Transcationid string `json:"transcationid"`
}

// for pm msg
type PM struct {
	Tz   string `json:"tz"`
	Wxid string `json:"wxid"`
	Msgs []*Msg `json:"msgs"`
}

// for init
type Tzs struct {
	Tzs    []Tz `json:"tzs"`
	Public bool `json:"public"`

	TzMap map[string]string `json:"-"` // wxid -> tz
}

type Tz struct {
	Tz     string   `json:"tz"`
	Wxid   string   `json:"wxid"`
	Groups []string `json:"groups"`
}
