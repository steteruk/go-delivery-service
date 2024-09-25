package kafka

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/linkedin/goavro"

	"github.com/IBM/sarama"
)

// Publisher Async send message in kafka.
type Publisher struct {
	producer             sarama.AsyncProducer
	topic                string
	schemaRegistryClient *CachedSchemaRegistryClient
}

// NewPublisher Create new Publisher Async for sending in kafka.
func NewPublisher(address []string, schemaRegistryServers []string, topic string) (*Publisher, error) {
	publisher := Publisher{}
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewManualPartitioner
	config.Producer.RequiredAcks = sarama.WaitForLocal
	producer, err := sarama.NewAsyncProducer(address, config)
	schemaRegistryClient := NewCachedSchemaRegistryClient(schemaRegistryServers)
	publisher.schemaRegistryClient = schemaRegistryClient

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

// GetSchemaId get schema id from schema-registry service.
func (publisher *Publisher) GetSchemaId(topic string, avroCodec *goavro.Codec) (int, error) {
	schemaId, err := publisher.schemaRegistryClient.CreateSubject(topic+"-value", avroCodec)
	if err != nil {
		return 0, err
	}
	return schemaId, nil
}

// PublishMessage  Send async message in kafka.
func (publisher *Publisher) PublishMessage(ctx context.Context, message []byte, key []byte, schema string) error {
	avroCodec, err := goavro.NewCodec(schema)
	if err != nil {
		return err
	}

	schemaId, err := publisher.GetSchemaId(publisher.topic, avroCodec)

	if err != nil {
		return err
	}

	if schemaId == 0 {

	}

	native, _, err := avroCodec.NativeFromTextual(message)
	if err != nil {
		return err
	}
	// Convert native Go form to binary Avro data
	binaryValue, err := avroCodec.BinaryFromNative(nil, native)
	if err != nil {
		return err
	}

	binaryMsg := &AvroEncoder{
		SchemaID: schemaId,
		Content:  binaryValue,
	}
	messageKafka := sarama.ProducerMessage{
		Topic: publisher.topic,
		Value: binaryMsg,
	}

	if key != nil {
		messageKafka.Key = sarama.StringEncoder(key)
	}

	publisher.publish(messageKafka)

	return nil
}

// AvroEncoder encodes schemaId and Avro message.
type AvroEncoder struct {
	SchemaID int
	Content  []byte
}

// Notice: the Confluent schema registry has special requirements for the Avro serialization rules,
// not only need to serialize the specific content, but also attach the Schema ID and Magic Byte.
// Ref: https://docs.confluent.io/current/schema-registry/serializer-formatter.html#wire-format
func (a *AvroEncoder) Encode() ([]byte, error) {
	var binaryMsg []byte
	// Confluent serialization format version number; currently always 0.
	binaryMsg = append(binaryMsg, byte(0))
	// 4-byte schema ID as returned by Schema Registry
	binarySchemaId := make([]byte, 4)
	binary.BigEndian.PutUint32(binarySchemaId, uint32(a.SchemaID))
	binaryMsg = append(binaryMsg, binarySchemaId...)
	// Avro serialized data in Avro's binary encoding
	binaryMsg = append(binaryMsg, a.Content...)
	return binaryMsg, nil
}

// Length of schemaId and Content.
func (a *AvroEncoder) Length() int {
	return 5 + len(a.Content)
}
