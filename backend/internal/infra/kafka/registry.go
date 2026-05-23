package kafka

import "errors"

// ProducerRegistry creates and owns all Kafka producers for the service.
// Adding a new producer requires only a new entry in Config.Producers — no wiring changes.
type ProducerRegistry struct {
	producers map[string]*Producer
}

func NewProducerRegistry(brokers []string, cfgs map[string]ProducerConfig) *ProducerRegistry {
	producers := make(map[string]*Producer, len(cfgs))
	for name, cfg := range cfgs {
		producers[name] = NewProducer(brokers, cfg)
	}
	return &ProducerRegistry{producers: producers}
}

// Get returns the producer registered under name. Returns nil if not found.
func (r *ProducerRegistry) Get(name string) *Producer {
	return r.producers[name]
}

// Close closes all producers and joins any errors.
func (r *ProducerRegistry) Close() error {
	var errs []error
	for _, p := range r.producers {
		if err := p.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
