package kafka

import (
	"strings"

	"github.com/suhrobdomoiZ/Eda-1/pkg/config"
)

type Config struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	ClientID      string
}

const (
	BrokersKey    config.Key = "KAFKA_BROKERS"
	TopicKey      config.Key = "KAFKA_TOPIC"
	ConsumerGroup config.Key = "KAFKA_CONSUMER_GROUP"
	ClientID      config.Key = "KAFKA_CLIENT_ID"
)

func Load() *Config {
	brokersStr := BrokersKey.MustGet()
	topic := TopicKey.MustGet()
	consumerGroup := ConsumerGroup.Get("")
	clientID := ClientID.MustGet()

	var brokers []string
	for _, b := range strings.Split(brokersStr, ",") {
		trimmed := strings.TrimSpace(b)
		if trimmed != "" {
			brokers = append(brokers, trimmed)
		}
	}

	return &Config{
		Brokers:       brokers,
		Topic:         topic,
		ConsumerGroup: consumerGroup,
		ClientID:      clientID,
	}
}
