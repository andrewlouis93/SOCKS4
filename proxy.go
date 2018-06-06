package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var SuccessfulSOCKConnectResponse = []byte{0x0, 0x5A, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
var FailedSOCKConnectResponse = []byte{0x0, 0x5B, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

// Source: https://en.wikipedia.org/wiki/SOCKS#SOCKS4
type SOCKRequest struct {
	version     int8
	commandCode int8
	port        string
	ipAddress   string
}

func marshalSOCKRequest(request []byte) (*SOCKRequest, error) {
	req := &SOCKRequest{
		version:     int8(request[0]),
		commandCode: int8(request[1]),
		port:        strconv.Itoa(int(binary.BigEndian.Uint16(request[2:4]))),
		ipAddress:   net.IP(request[4:8]).String(),
	}

	return req, nil
}

// TODO: We can handle errors better here.
func Handshake(connection net.Conn) (net.Conn, error) {
	buffer := make([]byte, 128)

	bytesRead, err := connection.Read(buffer)
	if err != nil {
		return nil, err
	}
	if bytesRead < 9 {
		return nil, fmt.Errorf("Less than 8 data bytes sent for SOCK connect request %v", buffer)
	}

	sReq, err := marshalSOCKRequest(buffer)
	if err != nil {
		return nil, err
	}

	if sReq.version != 4 {
		return nil, fmt.Errorf("Only SOCKS4 supported, you sent version: %v", sReq.version)
	}

	if sReq.commandCode != 1 {
		return nil, fmt.Errorf("Only Stream Connections 1 supported - you sent %v", sReq.commandCode)
	}

	addr := net.JoinHostPort(sReq.ipAddress, sReq.port)
	upstreamConn, err := net.Dial("tcp4", addr)
	if err != nil {
		connection.Write(FailedSOCKConnectResponse)
		upstreamConn.Close()
		return nil, err
	}

	connection.Write(SuccessfulSOCKConnectResponse)
	return upstreamConn, nil
}

// Pipe starts off goroutines streaming sockets to each other
func Pipe(conn1 net.Conn, conn2 net.Conn) chan error {
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

func handleClient(conn net.Conn) {
	upstreamConn, err := Handshake(conn)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Proxying: %v → %v\n", conn.RemoteAddr(), upstreamConn.RemoteAddr())
	errChan := Pipe(conn, upstreamConn)

	if err := <-errChan; err != nil {
		log.Println(err)
	}

	defer func() {
		conn.Close()
		upstreamConn.Close()
	}()

	return
}

func main() {
	sock, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Unable to start listening %v", err)
	}

	for {
		conn, err := sock.Accept()
		if err != nil {
			log.Println(err)
		} else {
			fmt.Printf("Accepted new connection: %v\n", conn.RemoteAddr())
			go handleClient(conn)
		}
	}
}
