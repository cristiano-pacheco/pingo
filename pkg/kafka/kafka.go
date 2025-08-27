package kafka

type Message struct {
	Topic   string
	Key     []byte
	Value   []byte
	Headers []Header
}

type Header struct {
	Key   string
	Value []byte
}

type Builder interface {
	BuildProducer(topic string) Producer
	BuildConsumer(topic string, groupID string) Consumer
}

type builder struct {
	config Config
}

func NewKafkaBuilder(config Config) Builder {
	return &builder{
		config: config,
	}
}
func (b *builder) BuildProducer(topic string) Producer {
	return newProducer(b.config, topic)
}

func (b *builder) BuildConsumer() Consumer {
	return NewConsumer(b.config)
}
