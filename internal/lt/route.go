package lt

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sjzar/chatlog/internal/errors"
	"github.com/sjzar/chatlog/internal/model"
)

// 群聊
func (s *Service) HandleLtGroupchat(c *gin.Context) {
	q := struct {
		Ltid     string `form:"ltid"`
		Chatroom string `form:"chatroom"`
	}{}
	if err := c.BindQuery(&q); err != nil {
		errors.Err(c, err)
		return
	}

	var msgs []*model.Message

	if q.Chatroom == "" {
		if q.Ltid == "" {
			errors.Err(c, errors.New(nil, 400, "ltid或chatroom不能为空"))
			return
		}
		// 不指定群聊，则读取此tz下所有群聊
		conf := s.ds.GetLtConfig(q.Ltid)
		for _, g := range conf.Tzs[0].Groups {

			msg, err := s.chatlogdb.GetMessages(time.Unix(g.Cursor, 0), time.Now(), g.Chatroom, "", "", 300, 0)
			if err != nil {
				errors.Err(c, err)
				return
			}
			msgs = append(msgs, msg...)
		}

	} else {
		// 读取指定群聊

	}

	c.JSON(http.StatusOK, msgs)

	// 获取联系人nickname sendname displayname

	// 组装
}

// 配置
func (s *Service) HandleLtConfig(c *gin.Context) {
	q := struct {
		Ltid string `form:"ltid"`
	}{}
	if err := c.BindQuery(&q); err != nil {
		errors.Err(c, err)
		return
	}

	conf := s.ds.GetLtConfig(q.Ltid)

	c.JSON(http.StatusOK, conf)
}

// for lt 成员
func (s *Service) HandleLtMembers(c *gin.Context) {
	q := struct {
		ltid string `form:"ltid"`
	}{}

	if err := c.BindQuery(&q); err != nil {
		errors.Err(c, err)
		return
	}

	s.ds.GetLtConfig(q.ltid)
}
