package lt

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sjzar/chatlog/internal/errors"
)

// for lt 群聊
func (s *Service) HandleLtGroupchat(c *gin.Context) {
	// 读取lt配置，获取需要读取的群聊列表
	// chatRooms, err := s.wechat.LoadLtGroupChats()
	// if err != nil {
	// 	errors.Err(c, err)
	// 	return
	// }
	// 获取联系人nickname sendname displayname

	// 组装

	q := struct {
		Ltid string `form:"ltid"`
	}{}
	if err := c.BindQuery(&q); err != nil {
		errors.Err(c, err)
		return
	}

	conf := s.ds.GetLtInfo(q.Ltid)

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

	s.ds.GetLtInfo(q.ltid)

}
