package network

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/eviltomorrow/open-terminal/lib/setting"
)

type IPType int

const (
	IPV4 IPType = iota
	IPV6
)

func GetInterfaceFirst() (string, error) {
	e := make([]error, 0, 2)
	ip, err := GetInterfaceIPv4First()
	if err != nil {
		e = append(e, err)
	}
	if ip != "" {
		return ip, nil
	}

	ip, err = GetInterfaceIPv6First()
	if err != nil {
		e = append(e, err)
	}
	if ip != "" {
		return ip, nil
	}

	return "", fmt.Errorf("panic: get ipv4/ipv6 failure, nest error: %v", errors.Join(e...))
}

func GetInterfaceIPv4First() (string, error) {
	return getInterfaceIPFirst(IPV4)
}

func GetInterfaceIPv6First() (string, error) {
	return getInterfaceIPFirst(IPV6)
}

func GetLocalareaIP(network, address string) (string, error) {
	conn, err := net.DialTimeout(network, address, setting.DEFUALT_HANDLE_10_SECOND)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	hostPort := strings.Split(localAddr.String(), ":")
	if len(hostPort) != 2 {
		return "", fmt.Errorf("panic: invalid host_port, value: %v", hostPort)
	}

	host := hostPort[0]
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")

	return host, nil
}

func getInterfaceIPFirst(it IPType) (string, error) {
	inters, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, inter := range inters {
		if inter.Flags&net.FlagUp != 0 && !strings.HasPrefix(inter.Name, "lo") && !strings.HasPrefix(inter.Name, "docker") && !strings.HasPrefix(inter.Name, "virbr") {
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					switch it {
					case IPV4:
						if ipnet.IP.To4() != nil {
							return ipnet.IP.String(), nil
						}
					case IPV6:
						if ipnet.IP.To16() != nil {
							return ipnet.IP.String(), nil
						}
					}
				}
			}
		}
	}
	return "", errors.New("panic: unable to get first ip")
}

func GetInterfaceIPList(filters ...func(string) bool) ([]string, error) {
	inters, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ipList := make([]string, 0, len(inters))
loop:
	for _, inter := range inters {
		for _, filter := range filters {
			if filter(inter.Name) {
				continue loop
			}
		}
		if inter.Flags&net.FlagUp != 0 && !strings.HasPrefix(inter.Name, "lo") {
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						ipList = append(ipList, ipnet.IP.String())
					} else if ipnet.IP.To16() != nil {
						ipList = append(ipList, ipnet.IP.String())
					}
				}
			}
		}
	}
	if len(ipList) == 0 {
		return nil, fmt.Errorf("not found any ip")
	}
	return ipList, nil
}
