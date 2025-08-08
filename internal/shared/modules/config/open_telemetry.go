package config

type Telemetry struct {
	Enabled      bool   `mapstructure:"TELEMETRY_TRACER_ENABLED"`
	TracerVendor string `mapstructure:"TELEMETRY_TRACER_VENDOR"`
	TracerURL    string `mapstructure:"TELEMETRY_TRACER_URL"`
}
