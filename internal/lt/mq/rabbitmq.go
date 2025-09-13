package mq

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sjzar/chatlog/internal/lt/model"
)

type RabbitMQSender struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRabbitMQSender(amqpURL, queueName string) (*RabbitMQSender, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &RabbitMQSender{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

func (s *RabbitMQSender) SendPad(pad *model.GM) error {
	body, err := json.Marshal(pad)
	if err != nil {
		return err
	}
	return s.channel.Publish(
		"",
		s.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (s *RabbitMQSender) Close() {
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}
