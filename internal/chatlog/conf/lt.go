package conf

type Lt struct {
	Tzs []Tz `json:"tzs"`
}

type Tz struct {
	Tz     string   `json:"tz"`
	Ltid   string   `json:"ltid"`
	Admins string   `json:"admins"`
	Groups []*Group `json:"groups"`
	Cron   string   `json:"cron"`
}

type Group struct {
	Name   string `json:"name"`
	Cursor int64  `json:"cursor"`
}
