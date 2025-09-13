package repository

import (
	"encoding/json"
	"strings"

	"github.com/sjzar/chatlog/internal/lt/model"
)

func (r *Repository) GetTzs4Lt() *model.Tzs {
	return r.tzs
}

func (r *Repository) initLtWatchList() error {
	// TODO init lt watch list
	body := []byte(`
		{
			"tzs": [
				{
					"tz": "tuanzi_chatlog",
					"wxid": "wxid_lw7htwzweu8e22,wxid_q7pbibmw8u8r22,guanjun915423",
					"groups": ["24967990639@chatroom", "44561777260@chatroom", "19400951536@chatroom", "45109606216@chatroom", "55855301186@chatroom"]
				}
			],
			"public": true
		}
	`)
	// wxid_ypg32n67jcil12 团子
	tzs := &model.Tzs{}
	if err := json.Unmarshal(body, &tzs); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}
	r.tzs = tzs

	tzs.TzMap = make(map[string]string)
	for _, tz := range tzs.Tzs {
		for _, wxid := range strings.Split(tz.Wxid, ",") {
			if wxid != "" && wxid != "null" {
				tzs.TzMap[wxid] = tz.Tz
			}
		}
		for _, group := range tz.Groups {
			tzs.TzMap[group] = tz.Tz
		}
		// TODO 加入缓存
		// for _, group := range tz.Groups {
		// 	r.chatRoomList = append(r.chatRoomList, group)
		// }
	}

	// r.watchList4Lt = append(r.watchList4Lt, "24967990639@chatroom", "44561777260@chatroom", "19400951536@chatroom",
	// 	"45109606216@chatroom", "55855301186@chatroom", "wxid_ypg32n67jcil12", "wxid_lw7htwzweu8e22", "wxid_q7pbibmw8u8r22")
	return nil
}
