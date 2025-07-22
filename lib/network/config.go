package network

import (
	"fmt"
	"net"

	jsoniter "github.com/json-iterator/go"
)

type Config struct {
	AccessIP   string `json:"access_ip" toml:"access_ip" mapstructure:"access_ip"`
	BindIP     string `json:"bind_ip" toml:"bind_ip" mapstructure:"bind_ip"`
	BindPort   int    `json:"bind_port" toml:"bind_port" mapstructure:"bind_port"`
	DisableTLS bool   `json:"disable_tls" toml:"-" mapstructure:"-"`
}

func (c *Config) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(c)
	return string(buf)
}

func (c *Config) VerifyConfig() error {
	if c.AccessIP != "" {
		ip := net.ParseIP(c.AccessIP)
		if ip == nil {
			return fmt.Errorf("grpc.access_ip has wrong format: %s", c.BindIP)
		}
	}
	if c.BindIP != "0.0.0.0" {
		ip := net.ParseIP(c.BindIP)
		if ip == nil {
			return fmt.Errorf("grpc.bind_ip has wrong format: %s", c.BindIP)
		}
	}

	if c.BindPort <= 0 || c.BindPort > 65535 {
		return fmt.Errorf("grpc.bind_port has wrong format: %d", c.BindPort)
	}

	if !c.DisableTLS {
		return fmt.Errorf("grpc.disable_tls must be true")
	}
	return nil
}
