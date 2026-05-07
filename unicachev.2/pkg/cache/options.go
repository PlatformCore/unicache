package cache

import "time"

type Options struct {
	Namespace            string
	DefaultTTL           time.Duration
	MaxEntries           int
	MaxBytes             int64
	Shards               int
	CleanupInterval      time.Duration
	StaleWhileRevalidate time.Duration
	NegativeTTL          time.Duration
	EnableSingleFlight   bool
	Metrics              Metrics
}

func DefaultOptions() Options {
	return Options{Namespace: "default", DefaultTTL: 5 * time.Minute, MaxEntries: 100_000, MaxBytes: 256 << 20, Shards: 64, CleanupInterval: 30 * time.Second, StaleWhileRevalidate: 30 * time.Second, EnableSingleFlight: true, Metrics: NoopMetrics{}}
}
func (o Options) normalize() Options {
	d := DefaultOptions()
	if o.Namespace == "" {
		o.Namespace = d.Namespace
	}
	if o.DefaultTTL == 0 {
		o.DefaultTTL = d.DefaultTTL
	}
	if o.MaxEntries <= 0 {
		o.MaxEntries = d.MaxEntries
	}
	if o.MaxBytes <= 0 {
		o.MaxBytes = d.MaxBytes
	}
	if o.Shards <= 0 {
		o.Shards = d.Shards
	}
	if o.CleanupInterval <= 0 {
		o.CleanupInterval = d.CleanupInterval
	}
	if o.Metrics == nil {
		o.Metrics = NoopMetrics{}
	}
	return o
}
