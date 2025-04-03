package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
	fmt.Println("Connection accepted")

	linesChan := getLinesChannel(conn)

	for line := range linesChan {
		fmt.Println(line)
	}

	defer conn.Close()
	defer lsnr.Close()
	fmt.Println("Connection closed")
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
		defer close(lines)
		currentLineContents := ""
		for {
			b := make([]byte, 8)
			n, err := f.Read(b)
			if err != nil {
				if currentLineContents != "" {
					lines <- currentLineContents
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			str := string(b[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- fmt.Sprintf("%s%s", currentLineContents, parts[i])
				currentLineContents = ""
			}
			currentLineContents += parts[len(parts)-1]
		}
	}()
	return lines
}
