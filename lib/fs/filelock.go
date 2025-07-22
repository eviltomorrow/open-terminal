package fs

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func CreateFlockFile(path string) (*os.File, error) {
	return createFlockFile(path)
}

func createFlockFile(flockFile string) (*os.File, error) {
	if err := MkdirAll(filepath.Dir(flockFile)); err != nil {
		return nil, err
	}
	flockF, err := os.Create(flockFile)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(flockF.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return nil, err
	}
	return flockF, nil
}
