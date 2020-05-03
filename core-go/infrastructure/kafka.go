package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"
)

// NewKafkaConsumer Obtain a new Apache Kafka consumer (subsciber)
// * Requires KAFKA_BROKERS OS env variable
func NewKafkaConsumer(ctx context.Context, consumerGroup, topic string) (*pubsub.Subscription, error) {
	consumerSub, err := pubsub.OpenSubscription(ctx, fmt.Sprintf(`kafka://%s?topic=%s`,
		strings.ToUpper(consumerGroup), strings.ToUpper(topic)))
	if err != nil {
		return nil, err
	}

	return consumerSub, nil
}

// NewKafkaProducer Obtain a new Apache Kafka producer (publisher)
// * Requires KAFKA_BROKERS OS env variable
func NewKafkaProducer(ctx context.Context, topic string) (*pubsub.Topic, error) {
	producer, err := pubsub.OpenTopic(ctx, fmt.Sprintf("kafka://%s", strings.ToUpper(topic)))
	if err != nil {
		return nil, err
	}

	return producer, nil
}
