package kafka

import (
	"github.com/Shopify/sarama"
)

const topic = "latest_position_courier"
const partition = 0

type CourierPublisher struct {
	producer sarama.AsyncProducer
}

func NewCourierPublisher() (*CourierPublisher, error) {
	brokers := []string{"localhost:9201"}
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &CourierPublisher{
		producer: producer,
	}, nil
}

func (cp *CourierPublisher) Publish(message string) error {
	cp.producer.Input() <- &sarama.ProducerMessage{
		Topic:     topic,
		Partition: partition,
		Value:     sarama.ByteEncoder(message),
	}

	if err := cp.producer.CommitTxn(); err != nil {
		return err
	}

	return nil
}
