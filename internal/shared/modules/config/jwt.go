package config

type JWT struct {
	PrivateKey          string `mapstructure:"JWT_PRIVATE_KEY"`
	Issuer              string `mapstructure:"JWT_ISSUER"`
	ExpirationInSeconds int64  `mapstructure:"JWT_EXPIRATION_IN_SECONDS"`
}
