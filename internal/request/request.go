package request

import (
	"fmt"
	"io"
	"strings"
)

const (
	bufferSize              = 8
	requestStateInitialized = 1
	requestStateDone        = 2
)

type Request struct {
	RequestLine RequestLine
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func PrintRequestLine(r *Request) {
	fmt.Println("Request line:")
	fmt.Println("- Method: " + r.RequestLine.Method)
	fmt.Println("- Target: " + r.RequestLine.RequestTarget)
	fmt.Println("- Version: " + r.RequestLine.HttpVersion)
}

func parseLineRequest(r *Request, data []byte) (bytesParsed int, err error) {
	//fmt.Printf("Buffer content stage 3: %q\n", data)
	request := string(data)
	//TODO: modify to only check 1 "\r\n" and do nothing with the rest*************************************************** maybe delete line 45
	//check if request has 4 "\r\n" in it
	if strings.Count(request, "\r\n") < 4 {
		return 0, nil
	}

	requestParts := strings.Split(request, "\r\n")
	if len(requestParts) == 0 {
		return 0, nil
	}

	requestLine := requestParts[0]
	if requestLine == "" {
		return 0, fmt.Errorf("request line is empty")
	}

	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid request line, must contain 3 parts")
	}
	// for i := 0; i < len(parts); i++ {
	// 	fmt.Println(parts[i])
	// }

	//check if Method has only capital letters
	for i := 0; i < len(parts[0]); i++ {
		if parts[0][i] < 'A' || parts[0][i] > 'Z' {
			return 0, fmt.Errorf("invalid method, must contain only capital letters")
		}
	}

	//check if HttpVersion is only "HTTP/1.1"
	if parts[2] != "HTTP/1.1" {
		return 0, fmt.Errorf("invalid HTTP Version, only HTTP/1.1 is supported")
	}

	r.RequestLine.Method = parts[0]
	r.RequestLine.RequestTarget = parts[1]
	r.RequestLine.HttpVersion = "1.1"
	r.state = requestStateDone
	bytesParsed = len(request)
	return bytesParsed, nil
}

func (r *Request) parse(data []byte) (int, error) {
	bytesParsed := 0
	var err error
	//fmt.Printf("Buffer content stage 2: %q\n", data)
	for {
		switch r.state {
		case requestStateInitialized: // "initialized" state
			bytesParsed, err = parseLineRequest(r, data)
			if err != nil {
				return bytesParsed, err
			}
			return bytesParsed, nil
		case requestStateDone: // "done" state
			return 0, fmt.Errorf("error: trying to read data in a done state")
		default: // unknown state
			return 0, fmt.Errorf("error: unknown state")
		}
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	readToIndex := 0
	var r = Request{
		RequestLine: RequestLine{},
		state:       requestStateInitialized,
	}
	buf := make([]byte, bufferSize)
	for r.state != requestStateDone {
		if readToIndex == len(buf) { // If the buffer is full
			newBuf := make([]byte, len(buf)*2) // Create a new slice that's twice the size
			copy(newBuf, buf)                  // Copy the old data into the new slice
			buf = newBuf                       // Replace the old buffer with the new one
		}
		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				r.state = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead
		//fmt.Printf("Buffer content stage 1: %q\n", buf)

		parsedBytes, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		// Remove parsed data from the buffer
		copy(buf, buf[parsedBytes:readToIndex])
		readToIndex -= parsedBytes
	}
	return &r, nil
}
