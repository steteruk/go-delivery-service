package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
)

type Publisher struct {
	producer sarama.AsyncProducer
	topic    string
}

func NewPublisher(address, topic string) (*Publisher, error) {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewManualPartitioner
	config.Producer.RequiredAcks = sarama.WaitForLocal
	producer, err := sarama.NewAsyncProducer([]string{address}, config)

	if err != nil {
		return nil, fmt.Errorf("failed to create a new sarama async producer: %w", err)
	}

	return &Publisher{producer: producer, topic: topic}, nil
}

func (p *Publisher) publish(message sarama.ProducerMessage) {
	p.producer.Input() <- &message
}

func (p *Publisher) PublishMessage(message []byte) {
	p.publish(sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(message),
	})
}
