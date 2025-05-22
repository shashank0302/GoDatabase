package btree

import (
	"fmt"
	"testing"
)

func TestBTree_Insert(t *testing.T) {
	tree := NewBTree()

	// Test basic insertion
	err := tree.Insert([]byte("key1"), []byte("value1"))
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}

	// Test duplicate key
	err = tree.Insert([]byte("key1"), []byte("value2"))
	if err == nil {
		t.Error("Expected error for duplicate key")
	}

	// Test key too large
	largeKey := make([]byte, BTREE_MAX_KEY_SIZE+1)
	err = tree.Insert(largeKey, []byte("value"))
	if err == nil {
		t.Error("Expected error for key too large")
	}

	// Test value too large
	largeValue := make([]byte, BTREE_MAX_VAL_SIZE+1)
	err = tree.Insert([]byte("key"), largeValue)
	if err == nil {
		t.Error("Expected error for value too large")
	}
}

func TestBTree_Get(t *testing.T) {
	tree := NewBTree()

	// Insert some test data
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range testData {
		err := tree.Insert([]byte(k), []byte(v))
		if err != nil {
			t.Errorf("Insert failed: %v", err)
		}
	}

	// Test retrieval
	for k, v := range testData {
		value, err := tree.Get([]byte(k))
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if string(value) != v {
			t.Errorf("Expected %s, got %s", v, string(value))
		}
	}

	// Test non-existent key
	_, err := tree.Get([]byte("nonexistent"))
	if err == nil {
		t.Error("Expected error for non-existent key")
	}
}

func TestBTree_Delete(t *testing.T) {
	tree := NewBTree()

	// Insert test data
	err := tree.Insert([]byte("key1"), []byte("value1"))
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}

	// Test deletion
	err = tree.Delete([]byte("key1"))
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = tree.Get([]byte("key1"))
	if err == nil {
		t.Error("Expected error for deleted key")
	}

	// Test deleting non-existent key
	err = tree.Delete([]byte("nonexistent"))
	if err == nil {
		t.Error("Expected error for deleting non-existent key")
	}
}

func TestBTree_Size(t *testing.T) {
	tree := NewBTree()

	// Test empty tree
	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}

	// Insert some data
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("k%d", i))
		val := []byte(fmt.Sprintf("v%d", i))
		err := tree.Insert(key, val)
		if err != nil {
			t.Errorf("Insert failed: %v", err)
		}
	}

	// Test size after insertion
	if tree.Size() != 10 {
		t.Errorf("Expected size 10, got %d", tree.Size())
	}

	// Delete some data
	for i := 0; i < 5; i++ {
		err := tree.Delete([]byte(fmt.Sprintf("k%d", i)))
		if err != nil {
			t.Errorf("Delete failed: %v", err)
		}
	}

	// Test size after deletion
	if tree.Size() != 5 {
		t.Errorf("Expected size 5, got %d", tree.Size())
	}
}

func TestBTree_Height(t *testing.T) {
	tree := NewBTree()

	// Test empty tree
	if tree.Height() != 0 {
		t.Errorf("Expected height 0, got %d", tree.Height())
	}

	// Insert data to force tree growth
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		val := []byte(fmt.Sprintf("val_%d", i))
		err := tree.Insert(key, val)
		if err != nil {
			t.Errorf("Insert failed: %v", err)
		}
	}

	// Test height after insertion
	height := tree.Height()
	if height <= 0 {
		t.Errorf("Expected height > 0, got %d", height)
	}
} 