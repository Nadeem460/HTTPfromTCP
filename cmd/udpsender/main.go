package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("could not listen: %s", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("could not dial: %s", err)
	}
	rdr := *bufio.NewReader(os.Stdin)
	for {
		fmt.Println(">")
		line, err := rdr.ReadString('\n')
		if err != nil {
			log.Fatalf("could not read: %s", err)
		}
		timeout, err := conn.Write([]byte(line))
		if err != nil {
			log.Fatalf("could not write: %s", err)
		}
		if timeout == 0 {
			fmt.Println("Connection closed")
			break
		}
	}

	defer conn.Close()
}
