package main

import (
	"flag"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	handler := func(w *response.Writer, req *request.Request) {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			w.WriteStatusLine(response.StatusBadRequest)
			data := response.PageData{
				Title:   "400 Bad Request",
				Heading: "Bad Request",
				Message: "Your request honestly kinda sucked.",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
		case "/myproblem":
			w.WriteStatusLine(response.StatusInternalServerError)
			data := response.PageData{
				Title:   "500 Internal Server Error",
				Heading: "Internal Server Error",
				Message: "Okay, you know what? This one is on me.",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
		default:
			w.WriteStatusLine(response.StatusOK)
			data := response.PageData{
				Title:   "200 OK",
				Heading: "Success!",
				Message: "Your request was an absolute banger.",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
		}
	}

	httpbinHandler := func(w *response.Writer, req *request.Request) {
		if !strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			w.WriteStatusLine(response.StatusBadRequest)
			data := response.PageData{
				Title:   "400 Bad Request",
				Heading: "Unsupported Request",
				Message: "Your request honestly kinda sucked! Only /httpbin/ requests are supported.",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
			return
		}
		// Handle the request here
		// Trim the "/httpbin/" prefix
		target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
		headers := headers.Headers{
			"Content-Type":      "text/html",
			"Transfer-Encoding": "chunked",
			"Connection":        "close",
		}
		// Get request to httpbin.org
		resp, err := http.Get("https://httpbin.org/" + target)
		if err != nil {
			w.WriteStatusLine(response.StatusInternalServerError)
			data := response.PageData{
				Title:   "500 Internal Server Error",
				Heading: "Internal Server Error",
				Message: "httpbin.org is unresponsive.",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
			return
		}

		// read the response from httpbin.org and store it in a 1024 bytes buffer
		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				data := response.PageData{
					Title:   "500 Internal Server Error",
					Heading: "Internal Server Error",
					Message: "Couldn't read httpbin.org's response!",
				}
				w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
				w.WriteBody(data)
				return
			}
			log.Println("Bytes Read: ", n) //watchout n is in hex
			// Write each chunk read
			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(headers)
			_, err = w.WriteChunkedBody(buf)
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				data := response.PageData{
					Title:   "500 Internal Server Error",
					Heading: "Internal Server Error",
					Message: "Couldn't write part of the body!",
				}
				w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
				w.WriteBody(data)
				return
			}
		}
		// Write the suffix for the chunked data
		_, err = w.WriteChunkedBodyDone()
		if err != nil {
			w.WriteStatusLine(response.StatusInternalServerError)
			data := response.PageData{
				Title:   "500 Internal Server Error",
				Heading: "Internal Server Error",
				Message: "Couldn't write CRLF to end the body!",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
			return
		}
	}

	//=============================================================
	//Use flag -t to enable the test handler
	useTestHandler := flag.Bool("t", false, "use test handler")
	flag.Parse()
	if *useTestHandler {
		server, err := server.Serve(port, handler)
		if err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
		defer server.Close()
	}
	//=============================================================

	server, err := server.Serve(port, httpbinHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
