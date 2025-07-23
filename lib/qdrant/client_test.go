package qdrant

import (
	"testing"
	"time"
)

func TestInitQdrant(t *testing.T) {
	shutdown, err := InitQdrant(&Config{
		StartupRetryPeriod: 3 * time.Second,
		StartupRetryTimes:  3,
		ConnectTimeout:     5 * time.Second,
		Host:               "localhost",
		Port:               6334,
	})
	if err != nil {
		t.Fatalf("InitQdrant failure, nest error: %v", err)
	}
	defer shutdown()
}
