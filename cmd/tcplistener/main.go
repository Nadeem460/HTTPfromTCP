package main

import (
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	lsnr, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Fatalf("could not listen: %s", err)
	}

	var conn net.Conn
	for {
		conn, err = lsnr.Accept()
		if err != nil {
			log.Fatalf("could not accept: %s", err)
		} else {
			break
		}
	}
	//fmt.Println("Connection accepted")

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Fatalf("could not read request: %s", err)
	}

	request.PrintRequestLine(req)

	defer conn.Close()
	defer lsnr.Close()
	//fmt.Println("Connection closed")
}
