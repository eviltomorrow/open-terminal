package conf

import (
	"github.com/eviltomorrow/open-terminal/lib/config"
	"github.com/eviltomorrow/open-terminal/lib/flagsutil"
	"github.com/eviltomorrow/open-terminal/lib/log"
	"github.com/eviltomorrow/open-terminal/lib/network"
	jsoniter "github.com/json-iterator/go"
)

type Config struct {
	Log  *log.Config     `json:"log" toml:"log" mapstructure:"log"`
	GRPC *network.Config `json:"grpc" toml:"grpc" mapstructure:"grpc"`
}

func (c *Config) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(c)
	return string(buf)
}

func ReadConfig(opts *flagsutil.Flags) (*Config, error) {
	c := InitializeDefaultConfig(opts)

	if err := config.ReadFile(c, opts.ConfigFile); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) IsConfigValid() error {
	for _, f := range []func() error{
		c.Log.VerifyConfig,
		c.GRPC.VerifyConfig,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func InitializeDefaultConfig(opts *flagsutil.Flags) *Config {
	return &Config{
		Log: &log.Config{
			Level:         "info",
			DisableStdlog: opts.DisableStdlog,
		},
		GRPC: &network.Config{
			AccessIP:   "",
			BindIP:     "0.0.0.0",
			BindPort:   50001,
			DisableTLS: true,
		},
	}
}
