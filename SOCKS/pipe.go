package socks4

import (
	"io"
	"net"
)

// pipe kicks off goroutines streaming sockets to each other
func pipe(conn1 net.Conn, conn2 net.Conn) chan error {
	errChan := make(chan error)

	go func() {
		_, err := io.Copy(conn1, conn2)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(conn2, conn1)
		errChan <- err
	}()

	return errChan
}
