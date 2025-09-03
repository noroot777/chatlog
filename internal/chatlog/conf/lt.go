package conf

type Tz struct {
	Tz     string   `json:"tz"`
	Admins []string `json:"admins"`
	Groups []string `json:"groups"`
	Cron   string   `json:"cron"`
}
