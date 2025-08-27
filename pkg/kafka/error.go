package kafka

import "errors"

var (
	ErrInvalidSASLMechanism = errors.New("invalid SASL mechanism")
	ErrInvalidKafkaAddress  = errors.New("invalid Kafka address")
	ErrInvalidUserName      = errors.New("invalid username")
	ErrInvalidPassword      = errors.New("invalid password")
)
