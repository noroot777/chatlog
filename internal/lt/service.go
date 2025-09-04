package lt

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/sjzar/chatlog/internal/chatlog/conf"
	"github.com/sjzar/chatlog/internal/chatlog/webhook"
	"github.com/sjzar/chatlog/pkg/util"
)

type Service struct {
	sc            *conf.ServerConfig
	db            *sql.DB
	lt            conf.Lt
	webhook       *webhook.Service
	webhookCancel context.CancelFunc
}

func NewService(sc *conf.ServerConfig) *Service {
	return &Service{
		sc:            sc,
		db:            nil,
		webhook:       nil,
		webhookCancel: nil,
	}
}

func (s *Service) Init() {
	// 检查本地是否有lt数据库
	dbPath := filepath.Join(util.DefaultWorkDir(""), "lt.db")
	s.createDB(dbPath)

	// 加载数据库

}

func (s *Service) createDB(dbPath string) {
	// s.loadConfig()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic("failed to open database: " + err.Error())
	}
	// defer db.Close()
	s.db = db

	schema1 := `
	CREATE TABLE IF NOT EXISTS tzs (
		ltid TEXT PRIMARY KEY,
		tz TEXT NOT NULL,
		admins TEXT NOT NULL,
		cron TEXT
	);
	`
	schema2 := `
	CREATE TABLE IF NOT EXISTS groups (
		ltid TEXT PRIMARY KEY,
		tz TEXT NOT NULL,
		chatroom TEXT NOT NULL,
		cursor INTEGER DEFAULT 0
	);
	`
	s.db.Exec(schema1)
	s.db.Exec(schema2)

	s.loadConfig()
}

func (s *Service) loadConfig() {
	// addr := s.sc.Addr
	// // 从addr地址获取返回值，组成conf.Lt
	// resp, err := http.Get(addr)
	// if err != nil {
	// 	panic("failed to fetch config: " + err.Error())
	// }
	// defer resp.Body.Close()

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	panic("failed to read response body: " + err.Error())
	// }

	body := []byte(`
	{
		"tzs": [
			{
				"tz": "xxx",
				"ltid": "tzid_in_lt",
				"admins": "xxx",
				"groups": [
					{
						"name": "xxx@chatroom",
						"cursor": 17777777
					},
					{
						"name": "xxx@chatroom",
						"cursor": 17777777
					}
				],
				"cron": "abc"
			}
		]
	}`)

	// 假设 conf.Lt 是一个结构体，需要根据实际结构定义
	var ltConf conf.Lt
	if err := json.Unmarshal(body, &ltConf); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}

	// 将 ltConf 赋值到全局或需要的位置
	s.lt = ltConf
}
