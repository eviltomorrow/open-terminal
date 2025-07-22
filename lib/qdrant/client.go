package qdrant

import (
	"context"
	"time"

	"github.com/eviltomorrow/open-terminal/lib/zlog"
	"github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"
)

var Client *qdrant.Client

func InitQdrant(c *Config) (func() error, error) {
	client, err := tryConnect(c)
	if err != nil {
		return nil, err
	}
	Client = client

	return func() error {
		if Client == nil {
			return nil
		}

		return Client.Close()
	}, nil
}

func tryConnect(c *Config) (*qdrant.Client, error) {
	i := 1
	for {
		client, err := buildQdrant(c)
		if err == nil {
			return client, nil
		}
		zlog.Error("connect to qdrant failure", zap.Error(err))
		i++
		if i > c.StartupRetryTimes {
			return nil, err
		}

		time.Sleep(time.Duration(c.StartupRetryPeriod))
	}
}

func buildQdrant(c *Config) (*qdrant.Client, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   c.Host,
		Port:   c.Port,
		APIKey: c.APIKey,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	reply, err := client.HealthCheck(ctx)
	if err != nil {
		return nil, err
	}
	_ = reply.GetVersion()

	return client, nil
}
