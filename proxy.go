package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	SOCKS4 "github.com/andrewlouis93/SOCKS4/SOCKS"
)

var (
	port = flag.Int("port", 8080, "port you want socks4 server to listen on :)")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
}

func main() {
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening on :%d\n", *port)

	for {
		if conn, err := server.Accept(); err != nil {
			log.Println(err)
		} else {
			fmt.Printf("New connection: %v\n", conn.RemoteAddr())
			go SOCKS4.Handle(conn)
		}
	}
}
