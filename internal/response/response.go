package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"net/http"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reasonPhrase string
	if statusCode == StatusOK || statusCode == StatusBadRequest || statusCode == StatusInternalServerError {
		reasonPhrase = http.StatusText(int(statusCode))
	} else {
		reasonPhrase = "" //TODO: MAY NEED TO CHANGE TO SPACE to follow the HTTP spec
	}

	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": strconv.Itoa(contentLen), // fmt.Sprintf("%d", contentLen) is generally prefered but strconv.Itoa is faster
		"Content-Type":   "text/plain",
		"Connection":     "close",
	}
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		if _, err := fmt.Fprintf(w, "%s: %s\r\n", key, value); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\r\n")
	return err
}
