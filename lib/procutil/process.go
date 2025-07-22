package procutil

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"syscall"

	"github.com/eviltomorrow/open-terminal/lib/buildinfo"
	"github.com/fatih/color"
)

var reg = regexp.MustCompile(`\s+`)

func FindProcessWithPid(pid int) (*os.Process, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return nil, err
	}
	return process, nil
}

func FindProcessWithPidFile(path string) (*os.Process, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pid, err := strconv.Atoi(reg.ReplaceAllString(string(buf), ""))
	if err != nil {
		return nil, err
	}
	return os.FindProcess(pid)
}

func StopProcessWithPidFile(path string) error {
	if process, err := FindProcessWithPidFile(path); err != nil {
		return fmt.Errorf("load process with pidfile failure, nest error: %v", err)
	} else {
		if err := process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("panic: notify process with signal SIGTERM failure, nest error: %v", err)
		}
		printStopped()
		return nil
	}
}

var (
	bold      = color.New(color.Bold)
	greenbold = color.New(color.FgGreen, color.Bold)
)

func printStopped() {
	fmt.Printf("%s %s \r\n",
		greenbold.Sprint("Status:"), bold.Sprint("stopped"),
	)
}

func printRunning(pid int) {
	fmt.Printf("%s %s\r\n%s %s, %s => [%s %s, %s %s/%s] \r\n",
		greenbold.Sprint("Status:"), bold.Sprint("running"),
		greenbold.Sprint("Version:"), bold.Sprint(buildinfo.MainVersion),
		greenbold.Sprint("Runtime"), greenbold.Sprint("Pid:"), bold.Sprintf("%d", pid),
		greenbold.Sprint("OS/Arch:"), bold.Sprint(runtime.GOOS), bold.Sprint(runtime.GOARCH),
	)
}
