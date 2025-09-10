package network

import (
	"fmt"
	"log"
	"net"
	
	"godatabase/internal/storage"
)

// Server represents a TCP server for the key-value store
type Server struct {
	addr    string
	storage storage.Storage
	ln      net.Listener
}

// NewServer creates a new TCP server
func NewServer(addr string, storage storage.Storage) *Server {
	return &Server{
		addr:    addr,
		storage: storage,
	}
}

// Start starts the TCP server
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.ln = ln
	
	log.Printf("Server listening on %s", s.addr)
	
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		
		go s.handleConnection(conn)
	}
}

// Stop stops the server
func (s *Server) Stop() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	log.Printf("New connection from %s", conn.RemoteAddr())
	
	for {
		// Read request
		msg, err := ReadMessage(conn)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Failed to read message: %v", err)
			}
			break
		}
		
		// Process request
		resp := s.processRequest(msg)
		
		// Send response
		if err := WriteResponse(conn, resp); err != nil {
			log.Printf("Failed to write response: %v", err)
			break
		}
	}
	
	log.Printf("Connection closed from %s", conn.RemoteAddr())
}

// processRequest processes a client request
func (s *Server) processRequest(msg *Message) *Response {
	switch msg.Op {
	case OpPut:
		return s.handlePut(msg.Key, msg.Value)
	case OpGet:
		return s.handleGet(msg.Key)
	case OpDelete:
		return s.handleDelete(msg.Key)
	default:
		return &Response{
			Status: StatusError,
			Error:  "invalid operation",
		}
	}
}

// handlePut handles a PUT request
func (s *Server) handlePut(key, value []byte) *Response {
	if err := s.storage.Put(key, value); err != nil {
		return &Response{
			Status: StatusError,
			Error:  err.Error(),
		}
	}
	
	return &Response{
		Status: StatusOK,
	}
}

// handleGet handles a GET request
func (s *Server) handleGet(key []byte) *Response {
	value, err := s.storage.Get(key)
	if err != nil {
		if err.Error() == "key not found" {
			return &Response{
				Status: StatusNotFound,
				Error:  err.Error(),
			}
		}
		return &Response{
			Status: StatusError,
			Error:  err.Error(),
		}
	}
	
	return &Response{
		Status: StatusOK,
		Value:  value,
	}
}

// handleDelete handles a DELETE request
func (s *Server) handleDelete(key []byte) *Response {
	if err := s.storage.Delete(key); err != nil {
		return &Response{
			Status: StatusError,
			Error:  err.Error(),
		}
	}
	
	return &Response{
		Status: StatusOK,
	}
} 