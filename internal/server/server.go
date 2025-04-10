package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"sync/atomic"
)

const (
	// State constants
	serverStateClosed      = -1
	serverStateInitialized = 0
)

type Server struct {
	listener net.Listener
	state    atomic.Int32
	closed   atomic.Bool
	handler  Handler
}

type HandlerError struct {
	Code    int
	Message string
}

func (e *HandlerError) Write(w io.Writer) {
	//fmt.Fprintf(w, "HTTP/1.1 %d\r\n%s", e.Code, e.Message)
	fmt.Fprintf(w, "%s", e.Message) //ONLY ERROR MESSAGE
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, h Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	srv := &Server{
		listener: listener,
		handler:  h,
	}
	srv.state.Store(serverStateInitialized)

	go srv.listen()

	return srv, nil
}

func (s *Server) Close() error {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return err
		}
		s.listener = nil
	}
	//s.state.Store(serverStateClosed)
	s.closed.Store(true)
	return nil
}

func (s *Server) listen() {
	for {
		if s.closed.Load() { // Maybe Unnecessary because we check it below
			// Listener is closed, exit the loop
			return
		}
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				// Listener is closed, exit the loop without logging an error
				return
			}
			log.Println("Error accepting connection:", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Parse the request from the connection
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("Error parsing request:", err)
		// Handle the error (e.g., send an error response)
		// conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	// Create a new response writer
	w := response.NewWriter(conn)

	// Call the handler with the response writer and request
	s.handler(w, req)

}
