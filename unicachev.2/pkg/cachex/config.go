package cachex

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Mode selects the cache backend created by NewFromConfig.
// Supported: memory, hybrid. Hybrid is memory-only in unicache and L1/L2/L3 in hybridcache.
type Mode string

const (
	ModeMemory Mode = "memory"
	ModeHybrid Mode = "hybrid"
)

// Config is intentionally plain Go so it can be filled from YAML/Viper/ENV/JSON
// in your microservice without forcing this shared lib to own config loading.
type Config struct {
	Enabled            bool          `json:"enabled" yaml:"enabled"`
	Mode               Mode          `json:"mode" yaml:"mode"`
	Namespace          string        `json:"namespace" yaml:"namespace"`
	DefaultTTL         time.Duration `json:"default_ttl" yaml:"default_ttl"`
	NegativeTTL        time.Duration `json:"negative_ttl" yaml:"negative_ttl"`
	StaleTTL           time.Duration `json:"stale_ttl" yaml:"stale_ttl"`
	MaxEntries         int           `json:"max_entries" yaml:"max_entries"`
	MaxBytes           int64         `json:"max_bytes" yaml:"max_bytes"`
	Shards             int           `json:"shards" yaml:"shards"`
	CleanupInterval    time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
	EnableSingleFlight bool          `json:"enable_singleflight" yaml:"enable_singleflight"`
	EnablePromotion    bool          `json:"enable_promotion" yaml:"enable_promotion"`
	DiskDir            string        `json:"disk_dir" yaml:"disk_dir"`
}

func DefaultConfig() Config {
	return Config{
		Enabled: true, Mode: ModeMemory, Namespace: "app", DefaultTTL: 5 * time.Minute,
		NegativeTTL: 10 * time.Second, StaleTTL: 30 * time.Second, MaxEntries: 100_000,
		MaxBytes: 256 << 20, Shards: 64, CleanupInterval: 30 * time.Second,
		EnableSingleFlight: true, EnablePromotion: true, DiskDir: "./.cache",
	}
}

func (c Config) Normalize() Config {
	d := DefaultConfig()
	if !c.Enabled {
		c.Enabled = d.Enabled
	}
	if c.Mode == "" {
		c.Mode = d.Mode
	}
	if c.Namespace == "" {
		c.Namespace = d.Namespace
	}
	if c.DefaultTTL <= 0 {
		c.DefaultTTL = d.DefaultTTL
	}
	if c.NegativeTTL < 0 {
		c.NegativeTTL = d.NegativeTTL
	}
	if c.StaleTTL < 0 {
		c.StaleTTL = d.StaleTTL
	}
	if c.MaxEntries <= 0 {
		c.MaxEntries = d.MaxEntries
	}
	if c.MaxBytes <= 0 {
		c.MaxBytes = d.MaxBytes
	}
	if c.Shards <= 0 {
		c.Shards = d.Shards
	}
	if c.CleanupInterval <= 0 {
		c.CleanupInterval = d.CleanupInterval
	}
	if c.DiskDir == "" {
		c.DiskDir = d.DiskDir
	}
	return c
}

// ConfigFromEnv reads optional CACHE_* variables. Use this for zero-boilerplate setup.
func ConfigFromEnv() Config {
	c := DefaultConfig()
	if v := os.Getenv("CACHE_ENABLED"); v != "" {
		c.Enabled = parseBool(v, c.Enabled)
	}
	if v := os.Getenv("CACHE_MODE"); v != "" {
		c.Mode = Mode(strings.ToLower(v))
	}
	if v := os.Getenv("CACHE_NAMESPACE"); v != "" {
		c.Namespace = v
	}
	if v := os.Getenv("CACHE_DEFAULT_TTL"); v != "" {
		c.DefaultTTL = parseDuration(v, c.DefaultTTL)
	}
	if v := os.Getenv("CACHE_NEGATIVE_TTL"); v != "" {
		c.NegativeTTL = parseDuration(v, c.NegativeTTL)
	}
	if v := os.Getenv("CACHE_STALE_TTL"); v != "" {
		c.StaleTTL = parseDuration(v, c.StaleTTL)
	}
	if v := os.Getenv("CACHE_MAX_ENTRIES"); v != "" {
		c.MaxEntries = parseInt(v, c.MaxEntries)
	}
	if v := os.Getenv("CACHE_MAX_BYTES"); v != "" {
		c.MaxBytes = int64(parseInt(v, int(c.MaxBytes)))
	}
	if v := os.Getenv("CACHE_SHARDS"); v != "" {
		c.Shards = parseInt(v, c.Shards)
	}
	if v := os.Getenv("CACHE_CLEANUP_INTERVAL"); v != "" {
		c.CleanupInterval = parseDuration(v, c.CleanupInterval)
	}
	if v := os.Getenv("CACHE_SINGLEFLIGHT"); v != "" {
		c.EnableSingleFlight = parseBool(v, c.EnableSingleFlight)
	}
	if v := os.Getenv("CACHE_PROMOTION"); v != "" {
		c.EnablePromotion = parseBool(v, c.EnablePromotion)
	}
	if v := os.Getenv("CACHE_DISK_DIR"); v != "" {
		c.DiskDir = v
	}
	return c.Normalize()
}

func parseBool(v string, fallback bool) bool {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
func parseInt(v string, fallback int) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
func parseDuration(v string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
