package config

// TelemetryConfig holds the settings for OpenTelemetry integration
// This mirrors the Simba settings structs but is framework-agnostic.
type TelemetryConfig struct {
	Enabled        bool
	Tracing        TracingConfig
	Metrics        MetricsConfig
	ServiceName    string
	ServiceVersion string
	Environment    string
}

type TracingConfig struct {
	Enabled      bool
	Exporter     string
	Endpoint     string
	Insecure     bool
	SamplingRate float64
}

type MetricsConfig struct {
	Enabled        bool
	Exporter       string
	Endpoint       string
	Insecure       bool
	ExportInterval int
}
