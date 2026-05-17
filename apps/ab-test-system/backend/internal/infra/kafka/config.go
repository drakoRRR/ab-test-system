package kafka

import "time"

type Config struct {
	Brokers   []string                  `yaml:"brokers"`
	Producers map[string]ProducerConfig `yaml:"producers"`
	Consumers map[string]ConsumerConfig `yaml:"consumers"`
}

type ProducerConfig struct {
	Topic string `yaml:"topic"`
}

type ConsumerConfig struct {
	Topic        string        `yaml:"topic"`
	GroupID      string        `yaml:"group_id"`
	BatchSize    int           `yaml:"batch_size"`
	FlushTimeout time.Duration `yaml:"flush_timeout"`
}

func (c *ConsumerConfig) withDefaults() ConsumerConfig {
	if c.BatchSize == 0 {
		c.BatchSize = 100
	}
	if c.FlushTimeout == 0 {
		c.FlushTimeout = 2 * time.Second
	}
	return *c
}
