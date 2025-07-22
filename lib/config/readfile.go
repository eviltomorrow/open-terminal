package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eviltomorrow/open-terminal/lib/system"
	"github.com/spf13/viper"
)

type Instance interface {
	ShouldVerify
}

type ShouldVerify interface {
	IsConfigValid() error
}

func ReadFile(c Instance, path string) error {
	findConfigFile := func(path string) (string, error) {
		for _, p := range []string{
			path,
			filepath.Join(system.Directory.EtcDir, "config.toml"),
		} {
			fi, err := os.Stat(p)
			if err == nil && !fi.IsDir() {
				return p, nil
			}
		}
		return "", fmt.Errorf("not found config file")
	}

	configFile, err := findConfigFile(path)
	if err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return err
	}
	if err := v.Unmarshal(c); err != nil {
		return err
	}

	return c.IsConfigValid()
}
