package lt

import (
	"database/sql"
	"encoding/json"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/mapstructure"

	"github.com/sjzar/chatlog/internal/chatlog/conf"
	"github.com/sjzar/chatlog/internal/lt/model"
	"github.com/sjzar/chatlog/pkg/util"
)

type Service struct {
	sc *conf.ServerConfig
	db *sql.DB
}

func NewService(sc *conf.ServerConfig) *Service {
	return &Service{
		sc: sc,
		db: nil,
	}
}

func (s *Service) Init() {
	// 检查本地是否有lt数据库
	s.checkDB()
	// 每次启动都从线上拉取最新配置，以线上为准
	s.loadConfig()
}

func (s *Service) checkDB() {
	dbPath := filepath.Join(util.DefaultWorkDir(""), "lt.db")
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
		ltid TEXT NOT NULL,
		tz TEXT NOT NULL,
		chatroom TEXT NOT NULL,
		cursor INTEGER DEFAULT 0,
		PRIMARY KEY(ltid, chatroom)
	);
	`
	s.db.Exec(schema1)
	s.db.Exec(schema2)
}

func (s *Service) loadConfig() {
	/*
		任务：每3s访问 /api/chatlog/task?ltid=xxx GET 获取任务列表
		任务包含：
			{"taskid":"xxx","task" = ["lt_gm":{}, "lt_pm":{}, "lt_all_groups":{}, "lt_members":{}, "lt_cron":{}, "lt_admins":{}, "lt_groups":{}, "lt_update":[], "lt_notify":{}]}
			每次只有一个任务，任务为空时body为空
			1. 上传群聊内容 (定时+任务)				"lt_gm": {"ltid":"xxx"}
			2. 上传私聊内容（转账信息） (定时+任务)		"lt_pm": {"ltid":"xxx"}
			3. 上传群聊列表 (任务)					"lt_all_groups": {"ltid":"xxx"}
			4. 上传群成员列表 (任务)				"lt_members": {"ltid":"xxx", "groups":"gruop1,group2..."}
			5. 接收定时任务表达式 (任务)			"lt_cron": {"ltid":"xxx", "cron":"xxx"}
			6. 接收管理员列表 (任务)				"lt_admins": {"ltid":"xxx","cron":"xxx,xxx..."} (全量)
			7. 接收群信息的增删 (任务)			"lt_groups": {"ltid":"xxx", "groups":[{name:"xxx@chatroom", cursor:12345678}, ...]} (全量) (cursor信息的更改也放到此接口，以线上为准)
			8. 更新tz及群信息，包含新增/删除/更新 (任务)		"lt_update": [*conf.Lt]   其他tz的更新信息要同时传给公用号
			9. 欠费/开始传输/停止传输等 (任务)			"lt_notify": {"ltid":"xxx", "error":"xxx"} 开始传输时一定包含cursor信息
		返回任务执行结果：
			上传 /api/chatlog/task/result?taskid=xxx&ltid=xxx&type=上面七种代号&status=0	POST {具体结果}
			status=0表示成功，此时body内容为具体结果；非0表示失败，此时body为error信息
			具体结果如下：
				1. 与以前相同
				2. 与以前相同
				3.
				4. 与以前相同
				5. 空
				6. 空
				7. 空
			error信息body结构如下：
				{"error":"具体错误信息"}

		其他单次主动调用：
			1. 获取初始化信息 /api/chatlog/config?ltid=xxx	GET	{tzs:[{tz:"xxx", admins:"xxx,xxx", groups:[{name:"xxx@chatroom", cursor:12345678}, ...], cron:"xxx"}]}
				如果是公用号，则返回所有tz信息[&conf.Lt]
	*/
	// addr := s.sc.Addr + "?ltid=" + s.sc.Ltid
	// // 从addr地址获取返回值，组成conf.Lt
	// http.Head(s.sc.Token)
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
				"tz": "xxx2",
				"ltid": "tzid_in_lt",
				"admins": "xxx",
				"groups": [
					{
						"name": "xxx1@chatroom",
						"cursor": 17777777
					},
					{
						"name": "xxx2@chatroom",
						"cursor": 17777777
					}
				],
				"cron": "abc2"
			}
		]
	}`)

	// 假设 conf.Lt 是一个结构体，需要根据实际结构定义
	var ltConf model.LtConfig
	if err := json.Unmarshal(body, &ltConf); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}

	// 将 ltConf 赋值到全局或需要的位置

	// save to lt.db
	for _, tz := range ltConf.Tzs {
		var tzs model.Tzs
		mapstructure.Decode(tz, &tzs)

		s.db.Exec("REPLACE INTO tzs (ltid, tz, admins, cron) VALUES (?, ?, ?, ?)", tzs.Ltid, tzs.Tz, tzs.Admins, tzs.Cron)

		for _, group := range tz.Groups {
			var gr model.Groups
			mapstructure.Decode(tz, &gr)
			mapstructure.Decode(group, &gr)
			// gr.Chatroom = group.Name

			s.db.Exec("REPLACE INTO groups (ltid, tz, chatroom, cursor) VALUES (?, ?, ?, ?)", gr.Ltid, gr.Tz, gr.Chatroom, gr.Cursor)
		}
	}
}
