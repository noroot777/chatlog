package model

type LtConfig struct {
	Tzs []ConfigItem `json:"tzs"`
}

type ConfigItem struct {
	Tz     string           `json:"tz"`
	Ltid   string           `json:"ltid"`
	Admins string           `json:"admins"`
	Groups []*ConfigGroupIn `json:"groups"`
	Cron   string           `json:"cron"`
}

type ConfigGroupIn struct {
	Name   string `json:"name"`
	Cursor int64  `json:"cursor"`
}
