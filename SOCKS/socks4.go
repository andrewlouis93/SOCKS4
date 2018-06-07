package socks4

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
)

var (
	successfulSOCKConnect = []byte{0x0, 0x5A, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	failedSOCKConnect     = []byte{0x0, 0x5B, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
)

// The client connects to the SOCKS server and sends a CONNECT request when
// it wants to establish a connection to an application server. The client
// includes in the request packet the IP address and the port number of the
// destination host, and userid, in the following format.

// 				+----+----+----+----+----+----+----+----+----+----+....+----+
// 				| VN | CD | DSTPORT |      DSTIP        | USERID       |NULL|
// 				+----+----+----+----+----+----+----+----+----+----+....+----+
//  # bytes:	   1    1      2              4           variable       1

// ConnectRequest should contain SOCKS4 CONNECT request body, see below
type ConnectRequest struct {
	version     int8
	commandCode int8
	port        string
	ipAddress   string
}

func marshal(request []byte) (*ConnectRequest, error) {
	if len(request) < 9 {
		return nil, fmt.Errorf("Insufficient data for connect request")
	}

	connReq := &ConnectRequest{
		version:     int8(request[0]),
		commandCode: int8(request[1]),
		port:        strconv.Itoa(int(binary.BigEndian.Uint16(request[2:4]))),
		ipAddress:   net.IP(request[4:8]).String(),
	}

	if connReq.version != 4 {
		return nil, fmt.Errorf("Only SOCKS4 supported, you sent: %v", connReq.version)
	} else if connReq.commandCode != 1 {
		return nil, fmt.Errorf("Only Stream Connections 1 supported - you sent %v", connReq.commandCode)
	}

	return connReq, nil
}

// A reply packet is sent to the client when this connection is established,
// or when the request is rejected or the operation fails.

// 				+----+----+----+----+----+----+----+----+
// 				| VN | CD | DSTPORT |      DSTIP        |
// 				+----+----+----+----+----+----+----+----+
//  # of bytes:	   1    1      2              4

// VN is the version of the reply code and should be 0. CD is the result
// code with one of the following values:

// 	0x5A: request granted
// 	0x5B: request rejected or failed

// handshake: carries out the SOCK4 handshake, returns proxy → upstream connection
func handshake(connection net.Conn) (net.Conn, error) {
	buffer := make([]byte, 128)
	numBytesRead, err := connection.Read(buffer)
	if err != nil {
		return nil, err
	}

	bytesRead := make([]byte, numBytesRead)
	copy(bytesRead, buffer[:numBytesRead])

	sockReq, err := marshal(bytesRead)
	if err != nil {
		return nil, err
	}

	addr := net.JoinHostPort(sockReq.ipAddress, sockReq.port)
	upstreamConn, err := net.Dial("tcp4", addr)
	if err != nil {
		defer func() {
			connection.Write(failedSOCKConnect)
			upstreamConn.Close()
		}()
		return nil, err
	}

	connection.Write(successfulSOCKConnect)
	return upstreamConn, nil
}

// Handle receives a client connection, and sets up a proxied SOCKS4 conn to its destination
func Handle(conn net.Conn) {
	upstreamConn, err := handshake(conn)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Proxying: %v → %v\n", conn.RemoteAddr(), upstreamConn.RemoteAddr())
	errChan := pipe(conn, upstreamConn)

	if err := <-errChan; err != nil {
		log.Println(err)
	}

	defer func() {
		conn.Close()
		upstreamConn.Close()
	}()
}
