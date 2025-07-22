package procutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eviltomorrow/open-terminal/lib/buildinfo"
	"github.com/eviltomorrow/open-terminal/lib/fs"
	"github.com/eviltomorrow/open-terminal/lib/system"
)

func CreatePidFile() (func() error, error) {
	path := filepath.Join(system.Directory.VarDir, fmt.Sprintf("run/%s.pid", buildinfo.AppName))
	file, err := fs.CreateFlockFile(path)
	if err != nil {
		return nil, err
	}

	file.WriteString(fmt.Sprintf("%d", os.Getpid()))
	if err := file.Sync(); err != nil {
		file.Close()
		return nil, err
	}

	return func() error {
		if file != nil {
			if err := file.Close(); err != nil {
				return err
			}
			return os.Remove(path)
		}
		return fmt.Errorf("panic: pid file is nil")
	}, nil
}
