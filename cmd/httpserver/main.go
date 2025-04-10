package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
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

	server, err := server.Serve(port, handler)
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
