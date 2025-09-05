package model

type Tzs struct {
	Ltid   string `mapstructure:"ltid"`
	Tz     string `mapstructure:"tz"`
	Admins string `mapstructure:"admins"`
	Cron   string `mapstructure:"cron"`
}

type Groups struct {
	Ltid     string `mapstructure:"ltid"`
	Tz       string `mapstructure:"tz"`
	Chatroom string `mapstructure:"name"` //TODO chatroom
	Cursor   int64  `mapstructure:"cursor"`
}
