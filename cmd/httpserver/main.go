package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
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
		// Create Headers
		chunkedHeaders := headers.Headers{
			"Content-Type":      "text/html",
			"Transfer-Encoding": "chunked",
			"Connection":        "close",
			"Trailer":           "X-Content-SHA256, X-Content-Length",
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

		totalBytesRead := 0
		// store the full response from httpbin.org in a bytes.Buffer{}
		httpbinResponseBody := &bytes.Buffer{}
		// read the response from httpbin.org and store it in a 1024 bytes buffer
		buf := make([]byte, 1024)
		for {
			bytesRead, err := resp.Body.Read(buf)
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
			// store read bytes in totalBytesRead
			totalBytesRead += bytesRead
			log.Println("Bytes Read: ", bytesRead) //watchout n is in hex
			// append buf to responseBody - full response buffer
			httpbinResponseBody.Write(buf[:bytesRead])
			// Write each chunk read
			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(chunkedHeaders)
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
		// Calculate the hash of the response body
		hash := sha256.Sum256(httpbinResponseBody.Bytes())
		// Create Trailer Headers
		trailers := headers.Headers{
			"X-Content-SHA256": fmt.Sprintf("%x", hash),
			"X-Content-Length": fmt.Sprintf("%d", totalBytesRead),
		}
		// Write Trailer Headers
		err = w.WriteTrailers(trailers)
		if err != nil {
			w.WriteStatusLine(response.StatusInternalServerError)
			data := response.PageData{
				Title:   "500 Internal Server Error",
				Heading: "Internal Server Error",
				Message: "Couldn't write trailers",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
			return
		}
	}

	videoHandler := func(w *response.Writer, req *request.Request) {
		if req.RequestLine.RequestTarget != "/video" {
			w.WriteStatusLine(response.StatusBadRequest)
			data := response.PageData{
				Title:   "400 Bad Request",
				Heading: "Unsupported Request",
				Message: "Your request honestly kinda sucked! Only /video requests are supported.",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
			return
		}

		// Respond with the assets/vim.mp4 video
		videoPath := "assets/vim.mp4"
		videoFile, err := os.Open(videoPath)
		if err != nil {
			w.WriteStatusLine(response.StatusInternalServerError)
			data := response.PageData{
				Title:   "500 Internal Server Error",
				Heading: "Internal Server Error",
				Message: "Couldn't open the video file.",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
			return
		}
		defer videoFile.Close()

		// Get the file info to determine the size
		videoInfo, err := videoFile.Stat()
		if err != nil {
			w.WriteStatusLine(response.StatusInternalServerError)
			data := response.PageData{
				Title:   "500 Internal Server Error",
				Heading: "Internal Server Error",
				Message: "Couldn't retrieve video file info.",
			}
			w.WriteHeaders(response.GetDefaultHeaders(data.ContentLength()))
			w.WriteBody(data)
			return
		}

		// Set headers for video response
		headers := headers.Headers{
			"Content-Type":   "video/mp4",
			"Content-Length": fmt.Sprintf("%d", videoInfo.Size()),
		}
		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(headers)

		// Stream the video file to the response
		_, err = io.Copy(w, videoFile)
		if err != nil {
			log.Printf("Error streaming video: %v", err)
		}
	}

	//========================== HANDLER SELECTION ===================================
	//Use flag -t to enable the test handler
	//Use flag -v to enable the video handler
	//Use no flag to enable the chunked encoding handler
	useTestHandler := flag.Bool("t", false, "use test handler")
	useVideoHandler := flag.Bool("v", false, "use video handler")
	flag.Parse()
	if *useTestHandler { // test handler
		server, err := server.Serve(port, handler)
		if err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
		defer server.Close()
		log.Println("Server started on port", port, "in Testing Mode")
	} else if *useVideoHandler { // video handler
		server, err := server.Serve(port, videoHandler)
		if err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
		defer server.Close()
		log.Println("Server started on port", port, "in Video Mode")
	} else { // chunked encoding handler
		server, err := server.Serve(port, httpbinHandler)
		if err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
		defer server.Close()
		log.Println("Server started on port", port, "in Chunked Encoding Mode")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
