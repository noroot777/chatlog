package mq

import (
	"context"
	"encoding/json"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/sjzar/chatlog/internal/lt/model"
)

type RocketMQSender struct {
	producer rocketmq.Producer
	topic    string
}

var (
	rocketMQSenderPool = make(map[string]*RocketMQSender)
)

func NewRocketMQSender(topic string) (*RocketMQSender, error) {
	if rocketMQSenderPool[topic] != nil {
		return rocketMQSenderPool[topic], nil
	}

	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{"172.16.88.52:9876"}),
		producer.WithRetry(2),
	)
	if err != nil {
		return nil, err
	}
	if err := p.Start(); err != nil {
		return nil, err
	}
	rocketMQSenderPool[topic] = &RocketMQSender{
		producer: p,
		topic:    topic,
	}
	return rocketMQSenderPool[topic], nil
}

func (s *RocketMQSender) SendGm(msg *model.GM) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	rocketMsg := primitive.NewMessage(s.topic, body)
	rocketMsg.WithTag("gm")
	_, err = s.producer.SendSync(context.Background(), rocketMsg)
	return err
}

func (s *RocketMQSender) SendMembers(msg *model.Members) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	rocketMsg := primitive.NewMessage(s.topic, body)
	rocketMsg.WithTag("member")
	_, err = s.producer.SendSync(context.Background(), rocketMsg)
	return err
}

func (s *RocketMQSender) SendProdMsg(msg *model.PM) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	rocketMsg := primitive.NewMessage(s.topic, body)
	rocketMsg.WithTag("prod")
	_, err = s.producer.SendSync(context.Background(), rocketMsg)
	return err
}

func (s *RocketMQSender) SendCashes(cashes *model.Cashes) error {
	body, err := json.Marshal(cashes)
	if err != nil {
		return err
	}
	rocketMsg := primitive.NewMessage(s.topic, body)
	rocketMsg.WithTag("cashes")
	_, err = s.producer.SendSync(context.Background(), rocketMsg)
	return err
}

func (s *RocketMQSender) Close() {
	if s.producer != nil {
		s.producer.Shutdown()
	}
}
