package websocket

import (
	"encoding/json"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewMQ(queueName string) (*MQ, error) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	return &MQ{conn: conn, channel: ch, queue: q}, nil
}

func (mq *MQ) Publish(message Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return mq.channel.Publish(
		"",            // exchange
		mq.queue.Name, // routing key (queue name)
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (mq *MQ) ConsumeMessages(hub *Hub) {
	msgs, err := mq.channel.Consume(
		mq.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to register consumer:", err)
	}

	go func() {
		for d := range msgs {
			var msg Message
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Println("Failed to decode message:", err)
				continue
			}
			log.Printf("Received message: %+v\n", msg)
			hub.broadcast <- msg
		}
	}()
}
