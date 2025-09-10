package client

import (
	"context"
	"fmt"
	"time"

	"godatabase/internal/rpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a client for the distributed key-value store
// It implements the storage.Storage interface
type Client struct {
	conn   *grpc.ClientConn
	client proto.StorageClient
}

// New creates a new client (alias for NewClient)
func New(addr string) (*Client, error) {
	return NewClient(addr)
}

// NewClient creates a new client
func NewClient(addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &Client{
		conn:   conn,
		client: proto.NewStorageClient(conn),
	}, nil
}

// Put stores a key-value pair
func (c *Client) Put(key, value []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.Put(ctx, &proto.PutRequest{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("put failed: %s", resp.Error)
	}

	return nil
}

// Get retrieves a value for a key
func (c *Client) Get(key []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.Get(ctx, &proto.GetRequest{
		Key: key,
	})
	if err != nil {
		return nil, err
	}

	if !resp.Found {
		return nil, fmt.Errorf("key not found: %s", resp.Error)
	}

	return resp.Value, nil
}

// Delete removes a key-value pair
func (c *Client) Delete(key []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.Delete(ctx, &proto.DeleteRequest{
		Key: key,
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("delete failed: %s", resp.Error)
	}

	return nil
}

// Close closes the connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Size returns the number of keys (not implemented for client)
func (c *Client) Size() int {
	// This would require a new protocol message
	// For now, return -1 to indicate not supported
	return -1
}
