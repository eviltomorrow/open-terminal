package procutil

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/eviltomorrow/open-terminal/lib/network"
	jsoniter "github.com/json-iterator/go"
)

var pingbackHost = "127.0.0.1"

type BootInfo struct {
	ChallengeKey []byte `json:"challenge_key"`
	ListenPort   int    `json:"listen_port"`
}

func (bi *BootInfo) Marshal() ([]byte, error) {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(bi)
}

func (bi *BootInfo) UnMarshal(buf []byte) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(buf, bi)
}

func RunAppInBackground(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("panic: invalid os.Args format")
	}

	var (
		name    = args[0]
		newArgs = func() []string {
			data := make([]string, 0, len(args)-2)
			// for _, arg := range args[1:] {
			// 	switch arg {
			// 	case "-d", "--daemon":

			// 	default:
			// 		data = append(data, arg)
			// 	}
			// }
			data = append(data, "--disable-stdlog")
			return data
		}()
	)

	cmd := exec.Command(name, newArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	port, err := network.GetAvailablePort()
	if err != nil {
		return fmt.Errorf("generate available port failure, nest error: %v", err)
	}

	address := net.JoinHostPort(pingbackHost, fmt.Sprintf("%d", port))
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen %s failure, nest error: %v", address, err)
	}
	defer ln.Close()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return err
	}
	go func() {
		bi := &BootInfo{
			ChallengeKey: key,
			ListenPort:   port,
		}
		buf, err := bi.Marshal()
		if err != nil {
			log.Fatalf("[F] Marshal boot-info failure, nest error: %v", err)
		}
		_, _ = stdin.Write(buf)

		stdin.Close()
	}()

	if err := cmd.Start(); err != nil {
		return err
	}

	success, exit := make(chan struct{}), make(chan error)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Fatalf("[F] Accept tcp data failure, nest error: %v", err)
				}
				break
			}
			err = handlePingbackConn(conn, key)
			if err == nil {
				close(success)
				break
			}
			log.Fatalf("[F] Parse pingback data failure, nest error: %v", err)
		}
	}()

	go func() {
		err := cmd.Wait()
		exit <- err
	}()

	select {
	case <-success:
		printRunning(cmd.Process.Pid)
	case err := <-exit:
		return fmt.Errorf("process exited with error: %v", err)
	}
	return nil
}

func StopDaemon() error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		return fmt.Errorf("panic: unknown pipe")
	}

	confirmationBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if len(confirmationBytes) == 0 {
		return nil
	}

	bi := &BootInfo{}
	if err := bi.UnMarshal(confirmationBytes); err != nil {
		return err
	}

	address := net.JoinHostPort(pingbackHost, fmt.Sprintf("%d", bi.ListenPort))
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("dialing confirmation address: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(bi.ChallengeKey))
	if err != nil {
		return fmt.Errorf("writing confirmation bytes to %s, nest error: %v", address, err)
	}

	return nil
}

func handlePingbackConn(conn net.Conn, expect []byte) error {
	defer conn.Close()
	key, err := io.ReadAll(io.LimitReader(conn, 32))
	if err != nil {
		return err
	}
	if !bytes.Equal(key, expect) {
		return fmt.Errorf("wrong challenge key: %x", key)
	}
	return nil
}
