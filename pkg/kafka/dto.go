package kafka

import "time"

type Message struct {
	// Topic indicates which topic this message was consumed from via Reader.
	//
	// When being used with Writer, this can be used to configure the topic if
	// not already specified on the writer itself.
	Topic string

	// Partition is read-only and MUST NOT be set when writing messages
	Partition     int
	Offset        int64
	HighWaterMark int64
	Key           []byte
	Value         []byte
	Headers       []Header

	// This field is used to hold arbitrary data you wish to include, so it
	// will be available when handle it on the Writer's `Completion` method,
	// this support the application can do any post operation on each message.
	WriterData interface{}

	// If not set at the creation, Time will be automatically set when
	// writing the message.
	Time time.Time
}

type Header struct {
	Key   string
	Value []byte
}
