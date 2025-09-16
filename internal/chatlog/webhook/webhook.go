package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"

	"github.com/sjzar/chatlog/internal/chatlog/conf"
	"github.com/sjzar/chatlog/internal/lt/model"
	"github.com/sjzar/chatlog/internal/lt/mq"
	"github.com/sjzar/chatlog/internal/wechatdb"
)

type Config interface {
	GetWebhook() *conf.Webhook
}

type Webhook interface {
	Do(event fsnotify.Event)
}

type Service struct {
	config *conf.Webhook
	hooks  map[string][]*conf.WebhookItem
}

func New(config Config) *Service {
	s := &Service{
		config: config.GetWebhook(),
	}

	if s.config == nil {
		return s
	}

	hooks := make(map[string][]*conf.WebhookItem)
	for _, item := range s.config.Items {
		if item.Disabled {
			continue
		}
		if item.Type == "" {
			item.Type = "message"
		}
		switch item.Type {
		case "message":
			if hooks["message"] == nil {
				hooks["message"] = make([]*conf.WebhookItem, 0)
			}
			hooks["message"] = append(hooks["message"], item)
		default:
			log.Error().Msgf("unknown webhook type: %s", item.Type)
		}
	}
	s.hooks = hooks

	return s
}

func (s *Service) GetHooks(ctx context.Context, db *wechatdb.DB) []*Group {

	if len(s.hooks) == 0 {
		return nil
	}

	groups := make([]*Group, 0)
	for group, items := range s.hooks {
		hooks := make([]Webhook, 0)
		for _, item := range items {
			hooks = append(hooks, NewMessageWebhook(item, db, s.config.Host))
		}
		groups = append(groups, NewGroup(ctx, group, hooks, s.config.DelayMs))
	}

	return groups
}

type Group struct {
	ctx     context.Context
	group   string
	hooks   []Webhook
	delayMs int64
	ch      chan fsnotify.Event
}

func NewGroup(ctx context.Context, group string, hooks []Webhook, delayMs int64) *Group {
	g := &Group{
		group:   group,
		hooks:   hooks,
		delayMs: delayMs,
		ctx:     ctx,
		ch:      make(chan fsnotify.Event, 1),
	}
	go g.loop()
	return g
}

func (g *Group) Callback(event fsnotify.Event) error {
	// skip remove event
	if !event.Op.Has(fsnotify.Create) {
		return nil
	}

	select {
	case g.ch <- event:
	default:
	}
	return nil
}

func (g *Group) Group() string {
	return g.group
}

func (g *Group) loop() {
	for {
		select {
		case event, ok := <-g.ch:
			if !ok {
				return
			}
			if g.delayMs > 0 {
				time.Sleep(time.Duration(g.delayMs) * time.Millisecond)
			}
			g.do(event)
		case <-g.ctx.Done():
			return
		}
	}
}

func (g *Group) do(event fsnotify.Event) {
	for _, hook := range g.hooks {
		go hook.Do(event)
	}
}

type MessageWebhook struct {
	host     string
	conf     *conf.WebhookItem
	client   *http.Client
	db       *wechatdb.DB
	lastTime time.Time
}

func NewMessageWebhook(conf *conf.WebhookItem, db *wechatdb.DB, host string) *MessageWebhook {
	m := &MessageWebhook{
		host:     host,
		conf:     conf,
		client:   &http.Client{Timeout: time.Second * 10},
		db:       db,
		lastTime: time.Now(),
	}
	return m
}

func (m *MessageWebhook) Do(event fsnotify.Event) {

	/*
		公用号逻辑：
		 第一步：启动时调用线上接口，获取需要监控的群列表及tz，加入缓存
		 第二步：在本方法中过滤，属于监控对象则往下走
		 第三步：区分群聊与私聊
		 第四步：群聊中组装用户消息及新成员消息，发送mq
		 第五步：私聊消息如果是tz，则组装商品信息，发送mq
		tz逻辑：
		 第一步：启动时调用线上接口，获取需要监控的群列表，加入缓存
		 第二步：查询所有群成员，加入缓存
		 第三步：在本方法中过滤，属于监控对象则往下走
		 第四步：区分群聊与私聊
		 第五步：群聊中组装用户消息及新成员消息，发送mq
		 第六步：私聊中组装转账消息，发送mq
	*/

	watchList := make([]string, 0)
	tzs := make([]string, 0)
	groups := make([]string, 0)
	var watchListStr string

	tzConf := m.db.GetTzs4Lt()
	for _, tz := range tzConf.Tzs {
		tzs = append(tzs, tz.Wxid)
		groups = append(groups, tz.Groups...)
	}

	watchList = append(watchList, groups...)
	if tzConf.Public {
		// 公用号需要把tz也加进去
		watchList = append(watchList, tzs...)
	}
	watchListStr = strings.Join(watchList, ",")

	// TODO: just for test
	// watchListStr += ",wxid_q7pbibmw8u8r22,guanjun915423"

	gms := make(map[string]*model.GM)               // key: chatroom wxid, value: gm
	chatrooms := make(map[string]*withNewMember, 0) // key: chatroom wxid, value: withNewMember
	prodMsgs := make(map[string]*model.PM)          // key: tz, value: prodMsg
	cashes := make(map[string]*model.Cashes)        // key: tz, value: cashes
	for _, tz := range tzConf.Tzs {
		for _, group := range tz.Groups {
			gm := &model.GM{Tz: tz.Tz, Room: "", Wxid: group}
			gms[group] = gm
		}
	}

	// TODO 改造一下这个方法内部的查询sql逻辑，如果太多talker，分批查询
	messages, err := m.db.GetMessages(m.lastTime, time.Now().Add(time.Minute*10), watchListStr, "", m.conf.Keyword, 0, 0)
	if err != nil {
		log.Error().Err(err).Msgf("get messages failed")
		return
	}

	if len(messages) == 0 {
		return
	}

	m.lastTime = messages[len(messages)-1].Time.Add(time.Second)

	for _, message := range messages {
		// 处理群聊信息
		if strings.Contains(message.Talker, "@chatroom") {
			// 处理群聊新成员
			if message.Type == 10000 {
				if _, ok := chatrooms[message.Talker]; ok {
					continue
				}
				// 不做新成员的增量更新，直接全量更新
				// 系统消息 正则匹配 "xx邀请"yy"加入了群聊"
				re := regexp.MustCompile(`.*邀请"(.+)"加入了群聊`)
				matches := re.FindStringSubmatch(message.Content)
				if len(matches) == 2 {
					// newMemberName := matches[1]
					room := message.Talker
					tz := tzConf.TzMap[room]
					chatrooms[room] = &withNewMember{
						tz:       tz,
						chatroom: room,
						name:     message.TalkerName,
					}
				}
				continue
			}

			// 处理群聊消息
			gm := gms[message.Talker]
			gm.Room = message.TalkerName
			if message.Sender != "" && message.SenderName != "系统消息" {
				res, err := m.db.GetContacts(message.Sender, 1, 0)
				if err != nil || len(res.Items) == 0 {
					log.Error().Err(err).Msgf("获取昵称失败 %s", message.Sender)
				} else {
					contact := res.Items[0]
					message.SenderName = contact.NickName
					message.Remark = contact.Remark
				}
			}
			message.Content = message.PlainTextContent4Lt()
			msg := &model.Msg{
				Wxid:     message.Sender,
				Nickname: message.SenderName,
				// Remark:   message.Remark,
				// Roomnickname: message.SenderName,
				When: message.Time.Format(time.DateTime),
				Msg:  message.Content,
			}
			gm.Msgs = append(gm.Msgs, msg)
		} else {
			if tzConf.Public {
				if message.IsSelf {
					// 如果是公用号自己的消息，跳过
					continue
				}
				// 公用号只处理tz的商品信息
				tz := tzConf.TzMap[message.Talker]
				if prodMsgs[tz] == nil {
					prodMsgs[tz] = &model.PM{Tz: tz, Msgs: make([]*model.Msg, 0)}
				}
				prodMsg := prodMsgs[tz]
				msg := &model.Msg{
					Wxid:     message.Sender,
					Nickname: message.SenderName, // TODO 这个字段如果有remark，会是remark值，修改
					// Remark:   message.Remark,
					When: message.Time.Format(time.DateTime),
					Msg:  message.PlainTextContent4Lt(),
				}
				prodMsg.Msgs = append(prodMsg.Msgs, msg)
			} else {
				// 转账信息还是别在这处理了，弄成定时的
				// tz只处理私聊转账信息
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
		}
	}

	sendGm(gms)

	sendMembers(m, chatrooms)

	sendProdMsg(prodMsgs)

	sendCashes(cashes)
}

func sendCashes(cashes map[string]*model.Cashes) {
	for _, cash := range cashes {
		sender, err := mq.NewRocketMQSender(cash.Tz)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ create rocketmq sender failed")
			return
		}
		err = sender.SendCashes(cash)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ send rocketmq message failed")
		}
		str, _ := json.Marshal(&cash)
		log.Info().Msgf("cash information: %s", str)

		// TODO for test
		sender, err = mq.NewRocketMQSender("tuanzi_chatlog_mongo")
		if err != nil {
			log.Error().Err(err).Msgf("严重！ create rocketmq sender failed")
			return
		}
		err = sender.SendCashes(cash)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ send rocketmq message failed")
		}
	}
}

func sendProdMsg(prodMsgs map[string]*model.PM) {
	for _, pm := range prodMsgs {
		sender, err := mq.NewRocketMQSender(pm.Tz)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ create rocketmq sender failed")
			return
		}
		err = sender.SendProdMsg(pm)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ send rocketmq message failed")
		}
		str, _ := json.Marshal(&pm)
		log.Info().Msgf("tz私聊信息: %s", str)

		// TODO for test
		sender, err = mq.NewRocketMQSender("tuanzi_chatlog_mongo")
		if err != nil {
			log.Error().Err(err).Msgf("严重！ create rocketmq sender failed")
			return
		}
		err = sender.SendProdMsg(pm)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ send rocketmq message failed")
		}
	}
}

func sendGm(gms map[string]*model.GM) {
	for _, gm := range gms {
		if len(gm.Msgs) == 0 {
			continue
		}
		sender, err := mq.NewRocketMQSender(gm.Tz)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ create rocketmq sender failed")
			return
		}
		err = sender.SendGm(gm)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ send rocketmq message failed")
		}
		str, _ := json.Marshal(&gm)
		log.Info().Msgf("group message: %s", str)

		// TODO for test
		sender, err = mq.NewRocketMQSender("tuanzi_chatlog_mongo")
		if err != nil {
			log.Error().Err(err).Msgf("严重！ create rocketmq sender failed")
			return
		}
		err = sender.SendGm(gm)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ send rocketmq message failed")
		}
	}
}

func sendMembers(m *MessageWebhook, chatrooms map[string]*withNewMember) {
	for chatroom, chatroomInfo := range chatrooms {
		members := &model.Members{
			Tz: chatroomInfo.tz,
			Groups: []*model.MembersInGroup{
				{
					Wxid:     chatroom,
					Roomname: chatroomInfo.name,
					Members:  make([]*model.Member, 0),
				},
			},
		}

		res, err := m.db.GetChatRooms(chatroom, 1, 0)
		if err != nil || len(res.Items) == 0 {
			log.Error().Err(err).Msgf("没有获取到群聊")
			continue
		}
		chatRoom := res.Items[0]

		for _, user := range chatRoom.Users {
			res, err := m.db.GetContacts(user.UserName, 1, 0)
			if err != nil || len(res.Items) == 0 {
				log.Error().Err(err).Msgf("获取昵称失败 %s", user.UserName)
			} else {
				contact := res.Items[0]
				members.Groups[0].Members = append(members.Groups[0].Members, &model.Member{
					Wxid:         user.UserName,
					Nickname:     contact.NickName,
					Remark:       contact.Remark,
					Roomnickname: user.DisplayName,
					// Describe:     user.Signature,
					// Lalels:       user.ContactLabels,
				})
			}
		}
		// 本来想只添加新成员，但是chatRoom.DisplayName2User为空，无法根据昵称获取wxid，作罢
		// for _, newMember := range newMembers {
		// 	newMemberWxid := chatRoom.DisplayName2User[newMember]

		sender, err := mq.NewRocketMQSender(chatroomInfo.tz)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ create rocketmq sender failed")
			return
		}
		err = sender.SendMembers(members)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ send rocketmq message failed")
		}

		str, _ := json.Marshal(&members)
		log.Info().Msgf("new members join: %s", str)

		// TODO for test
		sender, err = mq.NewRocketMQSender("tuanzi_chatlog_mongo")
		if err != nil {
			log.Error().Err(err).Msgf("严重！ create rocketmq sender failed")
			return
		}
		err = sender.SendMembers(members)
		if err != nil {
			log.Error().Err(err).Msgf("严重！ send rocketmq message failed")
		}
	}
}

type withNewMember struct {
	tz       string
	chatroom string
	name     string
}
