package network

import (
	"fmt"
	"net"
	"sync"
)

// Client represents a TCP client for the key-value store
type Client struct {
	addr string
	conn net.Conn
	mu   sync.Mutex
}

// NewClient creates a new TCP client
func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

// Connect connects to the server
func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.conn = conn
	return nil
}

// Close closes the connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Put stores a key-value pair
func (c *Client) Put(key, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	
	// Send request
	msg := &Message{
		Op:    OpPut,
		Key:   key,
		Value: value,
	}
	if err := WriteMessage(c.conn, msg); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	
	// Read response
	resp, err := ReadResponse(c.conn)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.Status != StatusOK {
		return fmt.Errorf("server error: %s", resp.Error)
	}
	
	return nil
}

// Get retrieves a value for a key
func (c *Client) Get(key []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}
	
	// Send request
	msg := &Message{
		Op:  OpGet,
		Key: key,
	}
	if err := WriteMessage(c.conn, msg); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	
	// Read response
	resp, err := ReadResponse(c.conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.Status == StatusNotFound {
		return nil, fmt.Errorf("key not found")
	}
	if resp.Status != StatusOK {
		return nil, fmt.Errorf("server error: %s", resp.Error)
	}
	
	return resp.Value, nil
}

// Delete removes a key-value pair
func (c *Client) Delete(key []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	
	// Send request
	msg := &Message{
		Op:  OpDelete,
		Key: key,
	}
	if err := WriteMessage(c.conn, msg); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	
	// Read response
	resp, err := ReadResponse(c.conn)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.Status != StatusOK {
		return fmt.Errorf("server error: %s", resp.Error)
	}
	
	return nil
} 