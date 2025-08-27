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
	BuildProducer() Producer
	BuildConsumer() Consumer
}

type builder struct {
	config Config
}

func NewKafkaBuilder(config Config) Builder {
	return &builder{
		config: config,
	}
}
func (b *builder) BuildProducer() Producer {
	return NewProducer(b.config)
}

func (b *builder) BuildConsumer() Consumer {
	return NewConsumer(b.config)
}
