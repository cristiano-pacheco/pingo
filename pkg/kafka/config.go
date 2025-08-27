package kafka

type Config struct {
	Address []string
}

func MustNewConfig(address []string) Config {
	config, err := NewConfig(address)
	if err != nil {
		panic(err)
	}
	return config
}

func NewConfig(address []string) (Config, error) {
	if len(address) == 0 {
		return Config{}, ErrInvalidKafkaAddress
	}

	return Config{
		Address: address,
	}, nil
}
