package request

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strings"
)

const (
	bufferSize                 = 8
	requestStateInitialized    = 1
	requestStateDone           = 2
	requestStateParsingHeaders = 3
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
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
	fmt.Println("Headers:")
	for key, value := range r.Headers {
		fmt.Printf("- %s: %s\n", key, value)
	}
	//fmt.Println("End of request")
	//fmt.Println("====================================")
}

func parseLineRequest(r *Request, data []byte) (bytesParsed int, err error) {

	request := string(data)
	//Check at least one "\r\n" and do nothing with the rest
	if strings.Count(request, "\r\n") < 1 {
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
	r.state = requestStateParsingHeaders
	bytesParsed = len(requestLine) + 2 // +2 for "\r\n"
	return bytesParsed, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != requestStateDone {
		bytesParsed, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		totalBytesParsed += bytesParsed
		if bytesParsed == 0 {
			return totalBytesParsed, nil
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	bytesParsed := 0
	var err error

	switch r.state {
	case requestStateInitialized: // "initialized" state
		bytesParsed, err = parseLineRequest(r, data)
		if err != nil {
			return bytesParsed, err
		}
		return bytesParsed, nil
	case requestStateParsingHeaders: // "parsing headers" state
		bytesParsed, done, err := r.Headers.Parse(data)
		if err != nil {
			return bytesParsed, err
		}
		if done {
			r.state = requestStateDone
			return bytesParsed, nil
		}
		return bytesParsed, nil
	default: // unknown state
		return 0, fmt.Errorf("error: unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	readToIndex := 0
	var r = Request{
		RequestLine: RequestLine{},
		Headers:     headers.NewHeaders(),
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
				r.state = requestStateDone // TODO: Maybe end of stream - doesnt mean request is done
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

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
