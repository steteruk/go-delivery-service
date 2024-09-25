package kafka

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/linkedin/goavro"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
)

// JSONMessageHandler Handler jso message that sending in kafka.
type JSONMessageHandler interface {
	HandleJSONMessage(ctx context.Context, message []byte) error
}

// Consumer represents a Sarama consumer group consumer.
type Consumer struct {
	keepRunning          bool
	topic                string
	jsonMessageHandler   JSONMessageHandler
	ready                chan bool
	consumerGroup        sarama.ConsumerGroup
	schemaRegistryClient *CachedSchemaRegistryClient
}

// NewConsumer Create new Consumer with specify consumer group.
func NewConsumer(
	jsonMessageHandler JSONMessageHandler,
	brokers string,
	verbose bool,
	oldest bool,
	assignor string,
	topic string,
	kafkaSchemaRegistryAddress []string,
) (*Consumer, error) {
	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()

	switch assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	default:
		err := fmt.Errorf("unrecognized consumer group partition assignor: %s", assignor)

		return nil, err
	}

	if oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	consumerGroup, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), topic, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create courier location consumer: %w", err)
	}
	schemaRegistryClient := NewCachedSchemaRegistryClient(kafkaSchemaRegistryAddress)

	return &Consumer{
		consumerGroup:        consumerGroup,
		keepRunning:          true,
		ready:                make(chan bool),
		topic:                topic,
		jsonMessageHandler:   jsonMessageHandler,
		schemaRegistryClient: schemaRegistryClient,
	}, nil
}

// ConsumeMessage Consume message in format json from kafka for specific consumer group.
func (consumer *Consumer) ConsumeMessage(ctx context.Context) error {
	consumptionIsPaused := false
	ctx, cancel := context.WithCancel(ctx)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := consumer.consumerGroup.Consume(ctx, strings.Split(consumer.topic, ","), consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}

			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				log.Panicf("Error from consumer: %v", ctx.Err())
				return
			}

			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for consumer.keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")

			consumer.keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")

			consumer.keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(consumer.consumerGroup, &consumptionIsPaused)
		}
	}
	cancel()
	waitGroup.Wait()

	if err := consumer.consumerGroup.Close(); err != nil {
		return fmt.Errorf("Error closing client: %w", err)
	}

	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return errors.New("message channel was closed")
			}

			log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			messageAvro, err := consumer.ProcessAvroMsg(message)
			if err != nil {
				return errors.New("message is not valid")
			}

			err = consumer.jsonMessageHandler.HandleJSONMessage(session.Context(), messageAvro)

			if err != nil {
				return err
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Println("Resuming consumption")
	} else {
		client.PauseAll()
		log.Println("Pausing consumption")
	}

	*isPaused = !*isPaused
}

// GetSchema get schema  from schema-registry service
func (consumer *Consumer) GetSchema(id int) (*goavro.Codec, error) {
	codec, err := consumer.schemaRegistryClient.GetSchema(id)
	if err != nil {
		return nil, err
	}
	return codec, nil
}

// ProcessAvroMsg decode value and prepare for unmarshal
func (consumer *Consumer) ProcessAvroMsg(m *sarama.ConsumerMessage) ([]byte, error) {
	schemaId := binary.BigEndian.Uint32(m.Value[1:5])
	codec, err := consumer.GetSchema(int(schemaId))
	if err != nil {
		return nil, err
	}
	// Convert binary Avro data back to native Go form
	native, _, err := codec.NativeFromBinary(m.Value[5:])
	if err != nil {
		return nil, err
	}

	// Convert native Go form to textual Avro data
	textual, err := codec.TextualFromNative(nil, native)

	if err != nil {
		return nil, err
	}
	msg := textual
	return msg, nil
}
