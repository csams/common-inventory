package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Config struct {
	*Options

	// this can be set manually for testing
	KafkaConfig *kafka.ConfigMap
}

type completedConfig struct {
	KafkaConfig *kafka.ConfigMap
}

type CompletedConfig struct {
	*completedConfig
}

func NewConfig(o *Options) *Config {
	return &Config{
		Options: o,
	}
}

func (c *Config) Complete() (CompletedConfig, error) {
	var config *kafka.ConfigMap

	if c.KafkaConfig != nil {
		config = c.KafkaConfig
	} else {
		for _, o := range c.Opts {
			config.SetKey(o.Name, o.Value)
		}
	}

	return CompletedConfig{&completedConfig{
		KafkaConfig: config,
	}}, nil
}
