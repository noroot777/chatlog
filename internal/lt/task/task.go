package task

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sjzar/chatlog/internal/lt/model"
	TaskType "github.com/sjzar/chatlog/internal/lt/task/type"
	"github.com/sjzar/chatlog/internal/wechatdb"
)

type Task struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Desc  string `json:"desc"`
	Tz    string `json:"tz"`
	Group string `json:"group,omitempty"`
	Start string `json:"start"`
	End   string `json:"end"`
}

func FetchAndExecuteOnlineTask(db *wechatdb.DB, tzs *model.Tzs) {
	for _, tz := range tzs.Tzs {
		if tz.Token == "" {
			log.Error().Msgf("严重！ tz %s has no token, skip fetching tasks", tz.Tz)
			continue
		}

		url := fmt.Sprintf("https://your-api-host/api/task?token=%s", tz.Token)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("failed to fetch online task: %v\n", err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("failed to read response body: %v\n", err)
			return
		}

		fmt.Printf("online task response: %s\n", string(body))

		var task Task
		if err := json.Unmarshal(body, &task); err != nil {
			fmt.Printf("failed to unmarshal task: %v\n", err)
			return
		}

		// Execute the task
		executeTask(task, db)
	}
}

func executeTask(task Task, db *wechatdb.DB) {
	switch task.Type {
	case TaskType.Cash:
		watchListStr := ""
		startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", task.Start)
		if err != nil {
			fmt.Printf("failed to parse start time: %v\n", err)
			return
		}
		endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", task.End)
		if err != nil {
			fmt.Printf("failed to parse end time: %v\n", err)
			return
		}
		messages, err := db.GetMessages(startTime, endTime, watchListStr, "", "", 0, 0)
		if err != nil {
			fmt.Printf("failed to get messages: %v\n", err)
			return
		}
		cashes := make(map[string]*model.Cashes) // key: tz, value: cashes

		for _, message := range messages {
			if message.Type == 49 && message.SubType == 2000 {
				// TODO 根据联系人获取tz
				tz := "tuanzi_chatlog"
				if cashes[tz] == nil {
					cashes[tz] = &model.Cashes{
						Tz:       tz,
						Wxid:     message.Talker,
						CashList: make([]*model.Cash, 0),
					}
				}

				payInfo := message.MediaMsg.App.WCPayInfo
				if payInfo != nil {
					cashes[tz].CashList = append(cashes[tz].CashList, &model.Cash{
						Amount:        payInfo.FeeDesc,
						Transcationid: payInfo.TranscationID,
						Memo:          payInfo.PayMemo,
						Type:          strconv.Itoa(payInfo.PaySubType),
						Who:           message.Sender,
						When:          message.Time.Format(time.DateTime),
					})
				}
			}
		}

		fmt.Printf("Executing print task: %s\n", task.Desc)
	default:
		fmt.Printf("Unknown task type: %s\n", task.Type)
	}
}
