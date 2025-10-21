package otel

type Option func(*Telemetry)

func WithLogLevel(logLevel string) Option {
	return func(t *Telemetry) {
		t.logLevel = parseLogLevel(logLevel)
	}
}

func WithSampleRate(rate float64) Option {
	return func(t *Telemetry) {
		t.sampleRate = rate
	}
}

func WithServiceVersion(serviceVersion string) Option {
	return func(t *Telemetry) {
		t.serviceVersion = serviceVersion
	}
}
