package config

type Kafka struct {
	Address []string `mapstructure:"KAFKA_ADDRESS"`
	SASL    SASL     `mapstructure:",squash"`
}

type SASL struct {
	IsEnabled bool   `mapstructure:"KAFKA_SASL_ENABLED"`
	Mechanism string `mapstructure:"KAFKA_MECHANISM"` // "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"
	Username  string `mapstructure:"KAFKA_USERNAME"`
	Password  string `mapstructure:"KAFKA_PASSWORD"`
	UseTLS    bool   `mapstructure:"KAFKA_USE_TLS"` // enforce TLS if broker requires it

}
