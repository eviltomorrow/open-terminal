package network

import (
	"fmt"
	"net"
)

func GetAvailablePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}

	listen, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	defer listen.Close()

	return listen.Addr().(*net.TCPAddr).Port, nil
}

func IsPortAvailable(port int) bool {
	address := fmt.Sprintf("%s:%d", "0.0.0.0", port)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listen.Close()

	return true
}
