package qdrant

import (
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type Config struct {
	StartupRetryPeriod time.Duration `json:"startup_retry_period" toml:"-" mapstructure:"-"`
	StartupRetryTimes  int           `json:"startup_retry_times" toml:"-" mapstructure:"-"`
	ConnectTimeout     time.Duration `json:"connect_timeout" toml:"-" mapstructure:"-"`

	Host   string `json:"host" toml:"host" mapstructure:"host"`
	Port   int    `json:"port" toml:"port" mapstructure:"port"`
	APIKey string `json:"api_key" toml:"api-key" mapstructure:"api-key"`
}

func (c *Config) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(c)
	return string(buf)
}

func (c *Config) VerifyConfig() error {
	if c.Host == "" {
		return fmt.Errorf("qdrant.host is nil")
	}
	if c.Port == 0 {
		return fmt.Errorf("qdrant.port is 0")
	}
	if c.APIKey == "" {
		return fmt.Errorf("qdrant.api-key is nil")
	}
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("qdrant.connect_timeout has no value")
	}
	if c.StartupRetryTimes <= 0 {
		return fmt.Errorf("qdrant.startup_retry_times has no value")
	}
	if c.StartupRetryPeriod <= 0 {
		return fmt.Errorf("qdrant.startup_retry_period has no value")
	}

	return nil
}
