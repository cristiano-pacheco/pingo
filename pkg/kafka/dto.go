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
