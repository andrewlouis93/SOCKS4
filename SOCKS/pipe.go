package socks4

import (
	"net"
)

// io.Copy has hardcoded read buffer size of 32kb
func createConnectionChannel(conn net.Conn, errChan chan error) chan []byte {
	connChan := make(chan []byte)

	go func() {
		buff := make([]byte, 512)

		for {
			numBytes, err := conn.Read(buff)
			if numBytes > 0 {
				res := make([]byte, numBytes)
				copy(res, buff[:numBytes])
				connChan <- res
			}

			if err != nil {
				errChan <- err
				connChan <- nil
				break
			}
		}
	}()

	return connChan
}

func createConnectionChannels(conn1 net.Conn, conn2 net.Conn) (chan []byte, chan []byte, chan error) {
	errChan := make(chan error)

	c1Chan := createConnectionChannel(conn1, errChan)
	c2Chan := createConnectionChannel(conn2, errChan)

	return c1Chan, c2Chan, errChan
}

// pipe kicks off goroutines streaming sockets to each other
func pipe(conn1 net.Conn, conn2 net.Conn) error {
	c1Chan, c2Chan, errChan := createConnectionChannels(conn1, conn2)

	for {
		select {
		case buff1 := <-c1Chan:
			if buff1 != nil {
				conn2.Write(buff1)
			} else {
				return <-errChan
			}
		case buff2 := <-c2Chan:
			if buff2 != nil {
				conn1.Write(buff2)
			} else {
				return <-errChan
			}
		}
	}
}
