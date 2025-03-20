package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// RabbitMQ представляет соединение с RabbitMQ.
type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// New создает новое соединение с RabbitMQ.
func New(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		Conn:    conn,
		Channel: channel,
	}, nil
}

// Close закрывает соединение с RabbitMQ.
func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}

// DeclareQueue объявляет очередь в RabbitMQ.
func (r *RabbitMQ) DeclareQueue(queueName string) (amqp.Queue, error) {
	queue, err := r.Channel.QueueDeclare(
		queueName, // Имя очереди
		true,      // Долговечность
		false,     // Удалять при неиспользовании
		false,     // Эксклюзивность
		false,     // Без ожидания
		nil,       // Аргументы
	)
	if err != nil {
		log.Printf("[DeclareQueue]: Ошибка при создании очереди %s: %v", queueName, err)
		return amqp.Queue{}, err
	}

	log.Printf("[DeclareQueue]: Очередь %s успешно создана", queueName)
	return queue, nil
}

// Publish отправляет сообщение в очередь.
func (r *RabbitMQ) Publish(queueName string, body []byte) error {
	return r.Channel.Publish(
		"",        // Обменник
		queueName, // Ключ маршрутизации
		false,     // Обязательно
		false,     // Немедленно
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
}

// Consume получает сообщения из очереди.
func (r *RabbitMQ) Consume(queueName string) (<-chan amqp.Delivery, error) {
	return r.Channel.Consume(
		queueName, // Имя очереди
		"",        // Имя потребителя
		true,      // Автоподтверждение
		false,     // Эксклюзивность
		false,     // Без локального ожидания
		false,     // Без ожидания
		nil,       // Аргументы
	)
}

// GetMessages возвращает массив сообщений из очереди.
func (r *RabbitMQ) GetMessages(queueName string, maxMessages int) ([]map[string]interface{}, error) {
	if err := r.EnsureConnection(); err != nil {
		return nil, err
	}

	var messages []map[string]interface{}

	for i := 0; i < maxMessages; i++ {
		msg, ok, err := r.Channel.Get(queueName, true)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении сообщения: %w", err)
		}
		if !ok {
			break
		}

		var messageData map[string]interface{}
		if err := json.Unmarshal(msg.Body, &messageData); err != nil {
			log.Printf("Ошибка декодирования сообщения: %v", err)
			continue
		}

		messages = append(messages, messageData)
	}

	return messages, nil
}

func (r *RabbitMQ) EnsureConnection() error {
	if r.Conn == nil || r.Conn.IsClosed() {
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			return fmt.Errorf("не удалось подключиться к RabbitMQ: %w", err)
		}
		r.Conn = conn
	}

	if r.Channel == nil {
		channel, err := r.Conn.Channel()
		if err != nil {
			return fmt.Errorf("не удалось открыть канал: %w", err)
		}
		r.Channel = channel
	}

	return nil
}
