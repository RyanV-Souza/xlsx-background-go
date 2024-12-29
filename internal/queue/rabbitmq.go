package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName = "xlsx_exchange"
	QueueName    = "xlsx_queue"
	RoutingKey   = "xlsx.generate"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type GenerateXLSXPayload struct {
	UserID    uint       `json:"userId"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		ExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	_, err = ch.QueueDeclare(
		QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		QueueName,
		RoutingKey,
		ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQ) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, payload *GenerateXLSXPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = r.channel.PublishWithContext(
		ctx,
		ExchangeName,
		RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (r *RabbitMQ) ConsumeMessages(handler func(payload *GenerateXLSXPayload) error) error {
	msgs, err := r.channel.Consume(
		QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for msg := range msgs {
		var payload GenerateXLSXPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			msg.Nack(false, false)
			continue
		}

		if err := handler(&payload); err != nil {
			msg.Nack(false, true)
			continue
		}

		msg.Ack(false)
	}

	return nil
}
