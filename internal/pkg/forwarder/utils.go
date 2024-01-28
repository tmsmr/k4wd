package forwarder

import (
	"net"
)

func randomLocalPort() (port int, err error) {
	a, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	l, _ := net.ListenTCP("tcp", a)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
