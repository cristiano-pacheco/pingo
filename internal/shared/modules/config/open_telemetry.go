package config

type OpenTelemetry struct {
	Enabled             bool    `mapstructure:"OPEN_TELEMETRY_TRACER_ENABLED"`
	TracerVendor        string  `mapstructure:"OPEN_TELEMETRY_TRACER_VENDOR"`
	TracerURL           string  `mapstructure:"OPEN_TELEMETRY_TRACER_URL"`
	BatchTimeoutSeconds int     `mapstructure:"OPEN_TELEMETRY_BATCH_TIMEOUT_SECONDS"`
	MaxBatchSize        int     `mapstructure:"OPEN_TELEMETRY_MAX_BATCH_SIZE"`
	Insecure            bool    `mapstructure:"OPEN_TELEMETRY_INSECURE"`
	SampleRate          float64 `mapstructure:"OPEN_TELEMETRY_SAMPLE_RATE"`
}
