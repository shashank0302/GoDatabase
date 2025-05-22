package storage

import (
	"os"
	"testing"
)

func TestStorageEngine_Basic(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "db-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Create a new storage engine
	engine, err := NewStorageEngine(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	// Test Put
	err = engine.Put([]byte("key1"), []byte("value1"))
	if err != nil {
		t.Errorf("Put failed: %v", err)
	}

	// Test Get
	value, err := engine.Get([]byte("key1"))
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if string(value) != "value1" {
		t.Errorf("Expected value1, got %s", string(value))
	}

	// Test Delete
	err = engine.Delete([]byte("key1"))
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = engine.Get([]byte("key1"))
	if err == nil {
		t.Error("Expected error for deleted key")
	}
}

func TestStorageEngine_Concurrent(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "db-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Create a new storage engine
	engine, err := NewStorageEngine(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	// Test concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(i int) {
			key := []byte{byte(i)}
			value := []byte{byte(i + 100)}

			// Put
			if err := engine.Put(key, value); err != nil {
				t.Errorf("Concurrent Put failed: %v", err)
			}

			// Get
			if val, err := engine.Get(key); err != nil {
				t.Errorf("Concurrent Get failed: %v", err)
			} else if val[0] != value[0] {
				t.Errorf("Concurrent Get returned wrong value: expected %d, got %d", value[0], val[0])
			}

			// Delete
			if err := engine.Delete(key); err != nil {
				t.Errorf("Concurrent Delete failed: %v", err)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestStorageEngine_FileFormat(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "db-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Create a new storage engine
	engine, err := NewStorageEngine(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	engine.Close()

	// Open the file and verify the header
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Read the header
	header := make([]byte, 8)
	if _, err := file.Read(header); err != nil {
		t.Fatal(err)
	}

	// Verify magic number and version
	magic := uint32(header[0])<<24 | uint32(header[1])<<16 | uint32(header[2])<<8 | uint32(header[3])
	version := uint32(header[4])<<24 | uint32(header[5])<<16 | uint32(header[6])<<8 | uint32(header[7])

	if magic != MAGIC {
		t.Errorf("Invalid magic number: expected %x, got %x", MAGIC, magic)
	}
	if version != VERSION {
		t.Errorf("Invalid version: expected %d, got %d", VERSION, version)
	}
}

func TestStorageEngine_Size(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "db-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Create a new storage engine
	engine, err := NewStorageEngine(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	// Test empty size
	if engine.Size() != 0 {
		t.Errorf("Expected size 0, got %d", engine.Size())
	}

	// Insert some data
	for i := 0; i < 10; i++ {
		err := engine.Put([]byte{byte(i)}, []byte{byte(i)})
		if err != nil {
			t.Errorf("Put failed: %v", err)
		}
	}

	// Test size after insertion
	if engine.Size() != 10 {
		t.Errorf("Expected size 10, got %d", engine.Size())
	}

	// Delete some data
	for i := 0; i < 5; i++ {
		err := engine.Delete([]byte{byte(i)})
		if err != nil {
			t.Errorf("Delete failed: %v", err)
		}
	}

	// Test size after deletion
	if engine.Size() != 5 {
		t.Errorf("Expected size 5, got %d", engine.Size())
	}
} 