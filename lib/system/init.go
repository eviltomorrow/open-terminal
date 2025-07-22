package system

import (
	"os"
	"path/filepath"
	"strings"
)

func LoadRuntime() error {
	executePath, err := os.Executable()
	if err != nil {
		return err
	}
	executePath, err = filepath.Abs(executePath)
	if err != nil {
		return err
	}

	Directory.BinDir = filepath.Dir(executePath)
	if !strings.HasPrefix(Directory.BinDir, "/bin") {
		Directory.RootDir = filepath.Dir(Directory.BinDir)
	} else {
		Directory.RootDir = Directory.BinDir
	}
	Directory.EtcDir = filepath.Join(Directory.RootDir, "/etc")
	Directory.UsrDir = filepath.Join(Directory.RootDir, "/usr")
	Directory.VarDir = filepath.Join(Directory.RootDir, "/var")
	Directory.LogDir = filepath.Join(Directory.VarDir, "/log")

	Process.Name = filepath.Base(executePath)
	Process.Args = os.Args[1:]
	Process.Pid = os.Getpid()
	Process.PPid = os.Getppid()

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	Machine.Hostname = hostname

	return nil
}
