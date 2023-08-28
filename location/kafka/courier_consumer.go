package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/steteruk/go-delivery-service/location/domain"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const groupCode = "latest_position_courier"

type CourierConsumer struct {
	consumerGroup             sarama.ConsumerGroup
	keepRunning               bool
	courierLocationRepository domain.CourierLocationRepositoryInterface
	ready                     chan bool
}

func NewCourierConsumer(
	courierLocationRepository domain.CourierLocationRepositoryInterface,
	addr string,
	verbose bool,
	oldest bool,
	assignor string,

) (*CourierConsumer, error) {
	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	brokers := []string{addr}
	config := sarama.NewConfig()

	switch assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	default:
		log.Panicf("Unrecognized consumer group partition assignor: %s", assignor)
	}

	if oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupCode, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create courier location consumer: %w", err)
	}

	return &CourierConsumer{
		consumerGroup:             consumerGroup,
		keepRunning:               true,
		courierLocationRepository: courierLocationRepository,
		ready:                     make(chan bool),
	}, nil
}

func (c *CourierConsumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.ready)
	return nil
}

func (c *CourierConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *CourierConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			var courierLocation domain.CourierLocation
			if err := json.Unmarshal(message.Value, &courierLocation); err != nil {
				log.Printf("Failed to unmarshal Kafka message into courier location struct: %v\n", err)
				session.MarkMessage(message, "")
				return nil
			}
			err := c.courierLocationRepository.SaveLatestCourierGeoPosition(session.Context(), &courierLocation)
			if err != nil {
				log.Printf("Failed to save a courier location in the repository: %v", err)
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *CourierConsumer) toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Println("Resuming consumption")
	} else {
		client.PauseAll()
		log.Println("Pausing consumption")
	}

	*isPaused = !*isPaused
}

func (c *CourierConsumer) ConsumeCourierLatestCourierGeoPositionMessage(ctx context.Context) error {
	consumptionIsPaused := false
	ctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.consumerGroup.Consume(ctx, strings.Split(topic, ","), c); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				log.Panicf("Error from consumer: %v", ctx.Err())
				return
			}
			c.ready = make(chan bool)
		}
	}()

	<-c.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for c.keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			c.keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			c.keepRunning = false
		case <-sigusr1:
			c.toggleConsumptionFlow(c.consumerGroup, &consumptionIsPaused)
		}
	}
	cancel()
	wg.Wait()
	if err := c.consumerGroup.Close(); err != nil {
		return fmt.Errorf("Error closing client: %w", err)
	}

	return nil
}
