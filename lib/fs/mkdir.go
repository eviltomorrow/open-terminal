package fs

import (
	"fmt"
	"os"
)

func MkdirAll(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return os.MkdirAll(dir, 0o755)
	}
	if !fi.IsDir() {
		return fmt.Errorf("already exist same file, path: %v", dir)
	}
	return syncPath(dir)
}

func syncPath(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := d.Sync(); err != nil {
		_ = d.Close()
		return err
	}
	return d.Close()
}
