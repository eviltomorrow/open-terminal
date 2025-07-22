package envutil

import (
	"fmt"
	"path/filepath"

	"github.com/eviltomorrow/open-terminal/lib/finalizer"
	"github.com/eviltomorrow/open-terminal/lib/log"
	"github.com/eviltomorrow/open-terminal/lib/network"
	"github.com/eviltomorrow/open-terminal/lib/system"
	"github.com/eviltomorrow/open-terminal/lib/zlog"
)

func InitLog(log *log.Config) error {
	global, prop, err := zlog.InitLogger(&zlog.Config{
		Level:  log.Level,
		Format: "json",
		File: zlog.FileLogConfig{
			Filename:    filepath.Join(system.Directory.LogDir, "data.log"),
			MaxSize:     100,
			MaxDays:     30,
			MaxBackups:  90,
			Compression: "gzip",
		},
		DisableStacktrace: true,
		DisableStdlog:     log.DisableStdlog,
	})
	if err != nil {
		return fmt.Errorf("init global log failure, nest error: %v", err)
	}
	zlog.ReplaceGlobals(global, prop)
	finalizer.RegisterCleanupFuncs(global.Sync)

	return nil
}

func InitNetwork(network *network.Config) error {
	system.Network.AccessIP = func() string {
		if network.AccessIP != "" {
			return network.AccessIP
		} else {
			return system.Network.BindIP
		}
	}()
	system.Network.BindIP = func() string {
		if network.BindIP != "0.0.0.0" {
			return network.BindIP
		} else {
			return system.Network.BindIP
		}
	}()

	return nil
}

func InitClientForGRPCWithTarget(target string, fs ...func(string) error) error {
	for _, f := range fs {
		if err := f(target); err != nil {
			return err
		}
	}
	return nil
}
