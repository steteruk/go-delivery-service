package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

// Publisher Async send message in kafka
type Publisher struct {
	producer sarama.AsyncProducer
	topic    string
}

// NewPublisher Create new Publisher Async for sending in kafka
func NewPublisher(address, topic string) (*Publisher, error) {
	publisher := Publisher{}
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewManualPartitioner
	config.Producer.RequiredAcks = sarama.WaitForLocal
	producer, err := sarama.NewAsyncProducer([]string{address}, config)

	if err != nil {
		return nil, fmt.Errorf("failed to create a new sarama async producer: %w", err)
	}

	publisher.producer = producer
	publisher.topic = topic

	return &publisher, nil
}

func (publisher *Publisher) publish(message sarama.ProducerMessage) {
	publisher.producer.Input() <- &message
}

// PublishMessage  Send async message in kafka
func (publisher *Publisher) PublishMessage(ctx context.Context, message []byte, key []byte) error {

	messageKafka := sarama.ProducerMessage{
		Topic: publisher.topic,
		Value: sarama.StringEncoder(message),
	}

	if key != nil {
		messageKafka.Key = sarama.StringEncoder(key)
	}

	publisher.publish(messageKafka)

	return nil
}
