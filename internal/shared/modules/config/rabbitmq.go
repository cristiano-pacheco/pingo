package config

type RabbitMQ struct {
	Host     string `mapstructure:"RABBITMQ_HOST"`
	Port     string `mapstructure:"RABBITMQ_PORT"`
	Username string `mapstructure:"RABBITMQ_USERNAME"`
	Password string `mapstructure:"RABBITMQ_PASSWORD"`
	VHost    string `mapstructure:"RABBITMQ_VHOST"`
}
