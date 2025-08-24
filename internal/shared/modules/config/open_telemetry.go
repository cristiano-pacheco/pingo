package config

type OpenTelemetry struct {
	Enabled             bool    `mapstructure:"TELEMETRY_TRACER_ENABLED"`
	TracerVendor        string  `mapstructure:"TELEMETRY_TRACER_VENDOR"`
	TracerURL           string  `mapstructure:"TELEMETRY_TRACER_URL"`
	BatchTimeoutSeconds int     `mapstructure:"TELEMETRY_BATCH_TIMEOUT_SECONDS"`
	MaxBatchSize        int     `mapstructure:"TELEMETRY_MAX_BATCH_SIZE"`
	Insecure            bool    `mapstructure:"TELEMETRY_INSECURE"`
	SampleRate          float64 `mapstructure:"TELEMETRY_SAMPLE_RATE"`
}
