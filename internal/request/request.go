package request

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

const (
	bufferSize                 = 4096
	requestStateInitialized    = 1
	requestStateParsingHeaders = 2
	requestStateParseingBody   = 3
	requestStateDone           = 4
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
	fmt.Println("Body:")
	fmt.Println(string(r.Body))
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
			r.state = requestStateParseingBody
			return bytesParsed, nil
		}
		return bytesParsed, nil
	case requestStateParseingBody: // "parsing body" state
		// Check if there is "contect-length" header
		contentLengthStr, ok := r.Headers.Get("content-length")
		if ok {
			// If there is, read the body until the content length is reached
			// Convert contentLength to int
			contentLength, err := strconv.Atoi(contentLengthStr)
			if err != nil {
				return bytesParsed, fmt.Errorf("invalid content-length value: %v", err)
			}

			// Check if the body is complete
			if len(data) < contentLength {
				return bytesParsed, nil
			} else if len(data) > contentLength {
				return bytesParsed, fmt.Errorf("error: body is greater than content-length value")
			}

			// If reached here, the body is equal to content-length value
			// Copy the body to the request and change the state to done
			r.Body = data[:contentLength]
			r.state = requestStateDone
			return bytesParsed + contentLength, nil
		} else {
			// If there is no content length, go to done state
			r.state = requestStateDone
			return bytesParsed + len(data), nil
		}
	case requestStateDone: // "done" state
		return 0, fmt.Errorf("error: request is already done")
	default: // unknown state
		return 0, fmt.Errorf("error: unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	readToIndex := 0
	var r = Request{
		RequestLine: RequestLine{},
		Headers:     headers.NewHeaders(),
		Body:        make([]byte, 0),
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
				if r.state != requestStateDone {
					return nil, fmt.Errorf("error: unexpected end of stream")
				}
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
