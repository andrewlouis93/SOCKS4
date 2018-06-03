package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
)

// Source: https://en.wikipedia.org/wiki/SOCKS#SOCKS4
type SOCKRequest struct {
	version     int8
	commandCode int8
	port        string
	ipAddress   string
}

// BytesToIPAddress converts array of bytes to an IP address
func BytesToIPAddress(addressBytes []byte) string {
	return net.IPv4(
		addressBytes[0],
		addressBytes[1],
		addressBytes[2],
		addressBytes[3],
	).String()
}

// field 1: null byte
// field 2: status, 1 byte:
// 0x5A = request granted
// 0x5B = request rejected or failed
// 0x5C = request failed because client is not running identd (or not reachable from the server)
// 0x5D = request failed because client's identd could not confirm the user ID string in the request
// field 3: 2 arbitrary bytes, which should be ignored
// field 4: 4 arbitrary bytes, which should be ignored
func generateSOCKResponse(granted bool) []byte {
	if granted {
		return []byte{
			0x0,
			0x5A,
			0x0, 0x0,
			0x0, 0x0, 0x0, 0x0,
		}
	}

	return []byte{
		0x0,
		0x5B,
		0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
	}
}

// TODO: Error more consistently. Don't use log.Fatalf - shortcircuits the idiom...
func marshalSOCKRequest(request []byte) (*SOCKRequest, error) {
	if len(request) != 8 {
		log.Fatalf("SOCK proxy request not receiving 8 bytes %v", request)
	}

	req := &SOCKRequest{
		version:     int8(request[0]),
		commandCode: int8(request[1]),
		port:        strconv.Itoa(int(binary.BigEndian.Uint16(request[2:4]))),
		ipAddress:   BytesToIPAddress(request[4:8]),
	}

	if req.version != 4 {
		log.Fatalf("Only SOCKS4 supported; you provided: %v", req.version)
	}

	return req, nil
}

func readSOCKRequest(connection *net.Conn) (*SOCKRequest, error) {
	buffer := make([]byte, 8)

	for {
		var err error
		if _, err = (*connection).Read(buffer); err != nil {
			return nil, err
		}

		var sReq *SOCKRequest
		if sReq, err = marshalSOCKRequest(buffer); err != nil {
			return nil, err
		}

		return sReq, nil
	}
}

func main() {
	var err error
	var sock net.Listener

	if sock, err = net.Listen("tcp", ":8080"); err != nil {
		log.Fatalf("Unable to start listeing %v", err)
	}

	for {
		var conn net.Conn
		fmt.Println("Waiting for new connection...")
		if conn, err = sock.Accept(); err != nil {
			log.Fatalf("Unable to accept connection %v", err)
		}
		fmt.Println("Accepted new connection...")

		// Parse request
		var sockReq *SOCKRequest
		var response []byte
		if sockReq, err = readSOCKRequest(&conn); err != nil {
			response = generateSOCKResponse(false)
			conn.Write(response)
		}

		// Connect to upstream
		// var destConn net.Conn
		if _, err = net.Dial("tcp", sockReq.ipAddress+":"+sockReq.port); err != nil {
			response = generateSOCKResponse(false)
			conn.Write(response)
		} else {
			response = generateSOCKResponse(true)
			conn.Write(response)

			// TODO: Pipe the shits
		}
	}
}
