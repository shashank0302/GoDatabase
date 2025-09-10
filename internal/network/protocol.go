package network

import (
	"encoding/binary"
	"errors"
	"io"
)

// Operation types
const (
	OpPut    = byte(1)
	OpGet    = byte(2)
	OpDelete = byte(3)
)

// Response codes
const (
	StatusOK       = byte(0)
	StatusError    = byte(1)
	StatusNotFound = byte(2)
)

// Message represents a request/response message
type Message struct {
	Op    byte   // Operation type
	Key   []byte // Key
	Value []byte // Value (for Put operations and Get responses)
}

// Response represents a server response
type Response struct {
	Status byte   // Status code
	Value  []byte // Value (for Get responses)
	Error  string // Error message (if any)
}

// WriteMessage writes a message to the writer
func WriteMessage(w io.Writer, msg *Message) error {
	// Format: [Op(1)] [KeyLen(4)] [Key] [ValueLen(4)] [Value]
	
	// Write operation
	if err := binary.Write(w, binary.BigEndian, msg.Op); err != nil {
		return err
	}
	
	// Write key length and key
	if err := binary.Write(w, binary.BigEndian, uint32(len(msg.Key))); err != nil {
		return err
	}
	if _, err := w.Write(msg.Key); err != nil {
		return err
	}
	
	// Write value length and value
	if err := binary.Write(w, binary.BigEndian, uint32(len(msg.Value))); err != nil {
		return err
	}
	if _, err := w.Write(msg.Value); err != nil {
		return err
	}
	
	return nil
}

// ReadMessage reads a message from the reader
func ReadMessage(r io.Reader) (*Message, error) {
	msg := &Message{}
	
	// Read operation
	if err := binary.Read(r, binary.BigEndian, &msg.Op); err != nil {
		return nil, err
	}
	
	// Read key length and key
	var keyLen uint32
	if err := binary.Read(r, binary.BigEndian, &keyLen); err != nil {
		return nil, err
	}
	if keyLen > 1024*1024 { // 1MB max key size
		return nil, errors.New("key too large")
	}
	msg.Key = make([]byte, keyLen)
	if _, err := io.ReadFull(r, msg.Key); err != nil {
		return nil, err
	}
	
	// Read value length and value
	var valueLen uint32
	if err := binary.Read(r, binary.BigEndian, &valueLen); err != nil {
		return nil, err
	}
	if valueLen > 10*1024*1024 { // 10MB max value size
		return nil, errors.New("value too large")
	}
	msg.Value = make([]byte, valueLen)
	if _, err := io.ReadFull(r, msg.Value); err != nil {
		return nil, err
	}
	
	return msg, nil
}

// WriteResponse writes a response to the writer
func WriteResponse(w io.Writer, resp *Response) error {
	// Format: [Status(1)] [ValueLen(4)] [Value] [ErrorLen(4)] [Error]
	
	// Write status
	if err := binary.Write(w, binary.BigEndian, resp.Status); err != nil {
		return err
	}
	
	// Write value length and value
	if err := binary.Write(w, binary.BigEndian, uint32(len(resp.Value))); err != nil {
		return err
	}
	if _, err := w.Write(resp.Value); err != nil {
		return err
	}
	
	// Write error length and error
	errorBytes := []byte(resp.Error)
	if err := binary.Write(w, binary.BigEndian, uint32(len(errorBytes))); err != nil {
		return err
	}
	if _, err := w.Write(errorBytes); err != nil {
		return err
	}
	
	return nil
}

// ReadResponse reads a response from the reader
func ReadResponse(r io.Reader) (*Response, error) {
	resp := &Response{}
	
	// Read status
	if err := binary.Read(r, binary.BigEndian, &resp.Status); err != nil {
		return nil, err
	}
	
	// Read value length and value
	var valueLen uint32
	if err := binary.Read(r, binary.BigEndian, &valueLen); err != nil {
		return nil, err
	}
	if valueLen > 10*1024*1024 { // 10MB max value size
		return nil, errors.New("value too large")
	}
	resp.Value = make([]byte, valueLen)
	if _, err := io.ReadFull(r, resp.Value); err != nil {
		return nil, err
	}
	
	// Read error length and error
	var errorLen uint32
	if err := binary.Read(r, binary.BigEndian, &errorLen); err != nil {
		return nil, err
	}
	if errorLen > 1024 { // 1KB max error message
		return nil, errors.New("error message too large")
	}
	errorBytes := make([]byte, errorLen)
	if _, err := io.ReadFull(r, errorBytes); err != nil {
		return nil, err
	}
	resp.Error = string(errorBytes)
	
	return resp, nil
} 