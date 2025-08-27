package config

import (
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Environment   string        `mapstructure:"ENVIRONMENT"`
	HTTPPort      uint          `mapstructure:"HTTP_PORT"`
	CORS          CORS          `mapstructure:",squash"`
	JWT           JWT           `mapstructure:",squash"`
	DB            DB            `mapstructure:",squash"`
	MAIL          MAIL          `mapstructure:",squash"`
	OpenTelemetry OpenTelemetry `mapstructure:",squash"`
	App           App           `mapstructure:",squash"`
	Log           Log           `mapstructure:",squash"`
	RabbitMQ      RabbitMQ      `mapstructure:",squash"`
	Redis         Redis         `mapstructure:",squash"`
	Kafka         Kafka         `mapstructure:",squash"`
}

const EnvProduction = "production"
const EnvDevelopment = "development"
const EnvStaging = "staging"

var _global Config

func Init() {
	v := viper.New()

	// Allow environment variables to override config file settings
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Configure to read from .env file
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	// Read the config file (must exist)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	//nolint:sloglint // this is a module
	slog.Info("Using config file", "file", v.ConfigFileUsed())

	// Unmarshal the config into our struct
	if err := v.Unmarshal(&_global); err != nil {
		//nolint:sloglint // this is a module
		slog.Error("Failed to unmarshal config", "error", err)
		panic(err)
	}
}

func GetConfig() Config {
	return _global
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
