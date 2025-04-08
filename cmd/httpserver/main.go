package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	handler := func(w io.Writer, req *request.Request) *server.HandlerError {
		// If request path is /yourproblem return a 400 and a message "Your problem is not my problem\n"
		// If request path is /myproblem return a 500 and a message "Woopsie, my bad\n"
		// Otherwise, it should just write the string "All good, frfr\n" to the response body.
		if req.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				Code:    int(response.StatusBadRequest),
				Message: "Your problem is not my problem\n",
			}
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				Code:    int(response.StatusInternalServerError),
				Message: "Woopsie, my bad\n",
			}
		} else {
			_, err := w.Write([]byte("All good, frfr\n"))
			if err != nil {
				log.Println("Error writing response:", err)
				return nil
			}
			return nil
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
