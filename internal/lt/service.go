package lt

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sjzar/chatlog/internal/chatlog/conf"
	repository "github.com/sjzar/chatlog/internal/lt/datasource/v4"
	"github.com/sjzar/chatlog/internal/lt/task"
	"github.com/sjzar/chatlog/internal/wechatdb"
)

type Service struct {
	sc        *conf.ServerConfig
	chatlogdb *wechatdb.DB

	ds *repository.Service
}

func NewService(sc *conf.ServerConfig, db *wechatdb.DB) *Service {
	s := &Service{
		sc:        sc,
		chatlogdb: db,
		ds:        &repository.Service{},
	}

	s.init()

	return s
}

func (s *Service) init() {
	// 检查本地是否有lt数据库
	s.ds.CheckDB()
	// 每次启动都从线上拉取最新配置，以线上为准
	s.loadConfig()
	// 启动获取任务的定时程序
	s.initScheduler()
}

func (s *Service) initScheduler() {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	scheduler, err := gocron.NewScheduler(gocron.WithLocation(loc), gocron.WithGlobalJobOptions())
	if err != nil {
		panic("failed to create scheduler: " + err.Error())
	}
	// 每10s去线上获取任务
	scheduler.NewJob(
		gocron.DurationJob(10*time.Second),
		gocron.NewTask(task.FetchAndExecuteOnlineTask, s.chatlogdb, s.chatlogdb.GetTzs4Lt()),
		gocron.WithName("fetchAndExecuteTask"),
		gocron.WithTags("lt_online_task_every_10s"),
	)
	// scheduler.Start() // TODO start
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

	// body := []byte(`
	// {
	// 	"tzs": [
	// 		{
	// 			"tz": "xxx2",
	// 			"ltid": "tzid_in_lt",
	// 			"admins": "xxx",
	// 			"groups": [
	// 				{
	// 					"chatroom": "xxx1@chatroom",
	// 					"cursor": 17777777
	// 				},
	// 				{
	// 					"chatroom": "xxx2@chatroom",
	// 					"cursor": 17777777
	// 				}
	// 			],
	// 			"cron": "abc2"
	// 		}
	// 	]
	// }`)

	// // 假设 conf.Lt 是一个结构体，需要根据实际结构定义
	// var ltConf model.LtConfig
	// if err := json.Unmarshal(body, &ltConf); err != nil {
	// 	panic("failed to unmarshal config: " + err.Error())
	// }

	// // save to lt.db
	// for _, tz := range ltConf.Tzs {
	// 	var tzs model.Tzs
	// 	mapstructure.Decode(tz, &tzs)

	// 	_, err := s.ds.Exec("REPLACE INTO tzs (ltid, tz, admins, cron) VALUES (?, ?, ?, ?)", tzs.Ltid, tzs.Tz, tzs.Admins, tzs.Cron)
	// 	if err != nil {
	// 		panic(fmt.Sprintf("更新表数据失败 tzs: %+v, err : %s", tzs, err.Error()))
	// 	}

	// 	for _, group := range tz.Groups {
	// 		var gr model.Groups
	// 		mapstructure.Decode(tz, &gr)
	// 		mapstructure.Decode(group, &gr)
	// 		// gr.Chatroom = group.Name

	// 		_, err := s.ds.Exec("REPLACE INTO groups (ltid, tz, chatroom, cursor) VALUES (?, ?, ?, ?)", gr.Ltid, gr.Tz, gr.Chatroom, gr.Cursor)
	// 		if err != nil {
	// 			panic(fmt.Sprintf("更新表数据失败 groups: %+v, err : %s", gr, err.Error()))
	// 		}
	// 	}
	// }
}

// 适用于上面的方法
// type Tzs struct {
// 	Ltid   string `mapstructure:"ltid"`
// 	Tz     string `mapstructure:"tz"`
// 	Admins string `mapstructure:"admins"`
// 	Cron   string `mapstructure:"cron"`
// }

// type Groups struct {
// 	Ltid     string `mapstructure:"ltid"`
// 	Tz       string `mapstructure:"tz"`
// 	Chatroom string `mapstructure:"chatroom"`
// 	Cursor   int64  `mapstructure:"cursor"`
// }
