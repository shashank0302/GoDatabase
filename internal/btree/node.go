// pseudo code
package btree

import (
	"errors"
	"fmt"
)

const (
	BNODE_NODE = 1 // internal node: use child pointers
	BNODE_LEAF = 2 // leaf node: use value storage
)

const (
	BTREE_PAGE_SIZE    = 4096
	BTREE_MAX_KEY_SIZE = 1000
	BTREE_MAX_VAL_SIZE = 3000
)

// Node represents a B+tree node that can be serialized to a fixed 4K page.
// The on-disk layout is:
//
//   | type (2B) | nkeys (2B) | pointers (nkeys×8B) | offsets (nkeys×2B) | key-values (variable) | unused |
// 
// In this structure:
//   - For a leaf node (typ == BNODE_LEAF), the pointers are unused and values are stored
//     as key-value pairs inside the data section.
//   - For an internal node (typ == BNODE_NODE), each key has an associated child pointer (as a page number),
//     and the value size in the key-value pair is 0.
type Node struct {
	// Header
	typ   uint16 // Node type: BNODE_NODE or BNODE_LEAF
	nkeys uint16 // Number of keys stored

	// For internal nodes only. For leaf nodes, this remains unused.
	pointers []uint64 // Each 8 bytes representing a child pointer (page number)

	// Offsets into 'data' for each key-value pair (except the first which always starts at 0).
	offsets []uint16 // Each offset is 2 bytes

	// Encoded key-value pairs:
	// Each pair is stored as:
	//   | key_size (2B) | val_size (2B) | key (key_size bytes) | val (val_size bytes) |
	// In an internal node, the val_size is 0.
	data []byte // Concatenated key-value pairs
}

// Track parent-child relationships with a map
var nodeRelationships = make(map[uint64]*Node)
var nextNodeID uint64 = 1

// NewNode creates a new node of the specified type.
func NewNode(typ uint16) *Node {
	return &Node{
		typ:      typ,
		nkeys:    0,
		pointers: make([]uint64, 0),
		offsets:  make([]uint16, 0),
		data:     make([]byte, 0),
	}
}

// Reset clears the node's data.
func (n *Node) Reset() {
	n.nkeys = 0
	n.pointers = n.pointers[:0]
	n.offsets = n.offsets[:0]
	n.data = n.data[:0]
}

// Serialize converts the node to a byte slice.
func (n *Node) Serialize() []byte {
	// Calculate the total size needed for the serialized node.
	size := 4 + len(n.pointers)*8 + len(n.offsets)*2 + len(n.data)
	buf := make([]byte, size)

	// Write the header (type and nkeys).
	buf[0] = byte(n.typ >> 8)
	buf[1] = byte(n.typ)
	buf[2] = byte(n.nkeys >> 8)
	buf[3] = byte(n.nkeys)

	// Write the pointers.
	offset := 4
	for _, ptr := range n.pointers {
		buf[offset] = byte(ptr >> 56)
		buf[offset+1] = byte(ptr >> 48)
		buf[offset+2] = byte(ptr >> 40)
		buf[offset+3] = byte(ptr >> 32)
		buf[offset+4] = byte(ptr >> 24)
		buf[offset+5] = byte(ptr >> 16)
		buf[offset+6] = byte(ptr >> 8)
		buf[offset+7] = byte(ptr)
		offset += 8
	}

	// Write the offsets.
	for _, off := range n.offsets {
		buf[offset] = byte(off >> 8)
		buf[offset+1] = byte(off)
		offset += 2
	}

	// Write the data.
	copy(buf[offset:], n.data)

	return buf
}

// Deserialize converts a byte slice back into a node.
func (n *Node) Deserialize(data []byte) error {
	if len(data) < 4 {
		return errors.New("data too short")
	}

	// Read the header (type and nkeys).
	n.typ = uint16(data[0])<<8 | uint16(data[1])
	n.nkeys = uint16(data[2])<<8 | uint16(data[3])

	// Read the pointers.
	offset := 4
	n.pointers = make([]uint64, n.nkeys)
	for i := uint16(0); i < n.nkeys; i++ {
		n.pointers[i] = uint64(data[offset])<<56 | uint64(data[offset+1])<<48 | uint64(data[offset+2])<<40 | uint64(data[offset+3])<<32 | uint64(data[offset+4])<<24 | uint64(data[offset+5])<<16 | uint64(data[offset+6])<<8 | uint64(data[offset+7])
		offset += 8
	}

	// Read the offsets.
	n.offsets = make([]uint16, n.nkeys)
	for i := uint16(0); i < n.nkeys; i++ {
		n.offsets[i] = uint16(data[offset])<<8 | uint16(data[offset+1])
		offset += 2
	}

	// Read the data.
	n.data = make([]byte, len(data)-offset)
	copy(n.data, data[offset:])

	return nil
}

// Split splits the node into two nodes and returns (rightNode, promotedKey).
// The promotedKey is the smallest key in the right node which will be pushed up to the parent.
func (n *Node) Split() (*Node, []byte) {
	if n.nkeys < 2 {
		return nil, nil // nothing to split
	}

	splitIdx := n.nkeys / 2 // integer division

	// Create right node of same type
	right := NewNode(n.typ)

	// Copy pointers (for internal nodes only)
	if n.typ == BNODE_NODE {
		right.pointers = append(right.pointers, n.pointers[splitIdx:]...)
		n.pointers = n.pointers[:splitIdx]
	}

	// Data slice start where right node entries begin
	startOffset := n.offsets[splitIdx]

	// Copy data section for right node
	right.data = append(right.data, n.data[startOffset:]...)

	// Build offsets for right node (relative to its data slice)
	right.offsets = make([]uint16, n.nkeys-splitIdx)
	for i := uint16(0); i < uint16(len(right.offsets)); i++ {
		right.offsets[i] = n.offsets[splitIdx+i] - startOffset
	}

	// Update key counts
	right.nkeys = n.nkeys - splitIdx
	n.nkeys = splitIdx

	// Trim left node's offsets and data
	n.offsets = n.offsets[:splitIdx]
	n.data = n.data[:startOffset]

	// Determine promoted key (first key in right node)
	var promotedKey []byte
	if len(right.offsets) > 0 {
		o := right.offsets[0]
		if int(o)+4 <= len(right.data) {
			kLen := uint16(right.data[o])<<8 | uint16(right.data[o+1])
			kStart := o + 4
			kEnd := kStart + kLen
			if int(kEnd) <= len(right.data) {
				promotedKey = right.data[kStart:kEnd]
			}
		}
	}

	return right, promotedKey
}

// Merge merges the node with another node.
func (n *Node) Merge(other *Node) error {
	// Ensure both nodes are of the same type.
	if n.typ != other.typ {
		return errors.New("cannot merge nodes of different types")
	}

	// Append the keys, pointers, offsets, and data from the other node.
	n.pointers = append(n.pointers, other.pointers...)
	n.offsets = append(n.offsets, other.offsets...)
	n.data = append(n.data, other.data...)
	n.nkeys += other.nkeys

	return nil
}

// Validate checks the node's integrity.
func (n *Node) Validate() error {
	// Check if the number of keys matches the number of pointers and offsets.
	if n.nkeys != uint16(len(n.pointers)) || n.nkeys != uint16(len(n.offsets)) {
		return errors.New("inconsistent number of keys, pointers, or offsets")
	}

	// Check if the offsets are valid.
	for i := uint16(0); i < n.nkeys; i++ {
		if i > 0 && n.offsets[i] <= n.offsets[i-1] {
			return errors.New("invalid offsets")
		}
	}

	return nil
}

// String returns a string representation of the node for debugging.
func (n *Node) String() string {
	return fmt.Sprintf("Node{typ: %d, nkeys: %d, pointers: %v, offsets: %v, data: %v}", n.typ, n.nkeys, n.pointers, n.offsets, n.data)
}

// Iterate iterates over the keys and values in the node.
func (n *Node) Iterate(f func(key, value []byte) error) error {
	for i := uint16(0); i < n.nkeys; i++ {
		if int(i) >= len(n.offsets) {
			continue
		}
		start := n.offsets[i]
		if int(start)+4 > len(n.data) {
			continue
		}
		keyLen := uint16(n.data[start])<<8 | uint16(n.data[start+1])
		valLen := uint16(n.data[start+2])<<8 | uint16(n.data[start+3])
		keyStart := start + 4
		keyEnd := keyStart + keyLen
		if int(keyEnd) > len(n.data) {
			continue
		}
		key := n.data[keyStart:keyEnd]

		var value []byte
		if n.typ == BNODE_LEAF {
			valStart := keyEnd
			valEnd := valStart + valLen
			if int(valEnd) > len(n.data) {
				continue
			}
			value = n.data[valStart:valEnd]
		}

		if err := f(key, value); err != nil {
			return err
		}
	}
	return nil
}

// Size returns the current size of the node in bytes.
func (n *Node) Size() int {
	return 4 + len(n.pointers)*8 + len(n.offsets)*2 + len(n.data)
}

// IsFull checks if the node is full.
func (n *Node) IsFull() bool {
	return n.Size() >= BTREE_PAGE_SIZE
}

// IsEmpty checks if the node is empty.
func (n *Node) IsEmpty() bool {
	return n.nkeys == 0
}

// keys returns a slice of all keys in the node (without values).
func (n *Node) keys() [][]byte {
	if n.nkeys == 0 {
		return [][]byte{}
	}
	keys := make([][]byte, n.nkeys)
	for i := uint16(0); i < n.nkeys; i++ {
		if int(i) >= len(n.offsets) {
			continue
		}
		start := n.offsets[i]
		if int(start)+4 > len(n.data) {
			continue
		}
		keyLen := uint16(n.data[start])<<8 | uint16(n.data[start+1])
		// valLen := uint16(n.data[start+2])<<8 | uint16(n.data[start+3]) // not needed here
		keyStart := start + 4
		keyEnd := keyStart + keyLen
		if int(keyEnd) > len(n.data) {
			continue
		}
		keys[i] = n.data[keyStart:keyEnd]
	}
	return keys
}

// getChild returns the child pointer at the given index.
func (n *Node) getChild(i int) *Node {
	if i >= len(n.pointers) {
		return nil // We shouldn't create new nodes here - should return nil
	}
	
	// Get the node ID stored in the pointer
	nodeID := n.pointers[i]
	
	// Check if we have this node in our relationships map
	if child, exists := nodeRelationships[nodeID]; exists {
		return child
	}
	
	// If we reach here, either the node doesn't exist or it's not loaded
	// In a real implementation, we would load the node from disk
	// For now, create a new node and track the relationship
	child := NewNode(BNODE_LEAF)
	
	// Only store if we have a valid nodeID (not 0)
	if nodeID > 0 {
		nodeRelationships[nodeID] = child
	}
	
	return child
}

// setChild sets the child pointer at the given index.
func (n *Node) setChild(i int, child *Node) {
	// Ensure we have enough pointers
	if i >= len(n.pointers) {
		n.pointers = append(n.pointers, make([]uint64, i-len(n.pointers)+1)...)
	}
	
	// If the child node doesn't have an ID yet, assign one
	var nodeID uint64
	
	// Find the ID for this child node
	for id, node := range nodeRelationships {
		if node == child {
			nodeID = id
			break
		}
	}
	
	// If no existing ID found, create a new one
	if nodeID == 0 && child != nil {
		nodeID = nextNodeID
		nextNodeID++
		nodeRelationships[nodeID] = child
	}
	
	// Store the nodeID in the pointer
	n.pointers[i] = nodeID
}

// insertKV inserts a key-value pair at the given position.
func (n *Node) insertKV(pos int, key, value []byte) {
	// Encode the entry as |keyLen(2B)|valLen(2B)|key|value|
	keyLen := uint16(len(key))
	valLen := uint16(len(value))
	entrySize := 4 + int(keyLen) + int(valLen)
	entry := make([]byte, entrySize)
	// big-endian lengths
	entry[0] = byte(keyLen >> 8)
	entry[1] = byte(keyLen)
	entry[2] = byte(valLen >> 8)
	entry[3] = byte(valLen)
	copy(entry[4:], key)
	copy(entry[4+keyLen:], value)

	// Determine the byte offset where this entry should be inserted
	var newOffset uint16
	if pos == int(n.nkeys) {
		// append at end
		newOffset = uint16(len(n.data))
		n.data = append(n.data, entry...)
	} else {
		// insert in the middle, need to shift existing data to the right
		newOffset = n.offsets[pos]
		n.data = append(n.data[:newOffset], append(entry, n.data[newOffset:]...)...)
	}

	// Update offsets slice: insert newOffset at position pos
	n.offsets = append(n.offsets, 0)              // grow slice
	copy(n.offsets[pos+1:], n.offsets[pos:])      // shift right
	n.offsets[pos] = newOffset

	// Increment subsequent offsets to account for inserted bytes
	for i := pos + 1; i < len(n.offsets); i++ {
		n.offsets[i] += uint16(entrySize)
	}

	n.nkeys++
}

// getValue returns the value associated with key index i (for leaf nodes).
func (n *Node) getValue(i int) []byte {
	if n.typ != BNODE_LEAF || i < 0 || i >= int(n.nkeys) {
		return nil
	}
	start := n.offsets[i]
	if int(start)+4 > len(n.data) {
		return nil
	}
	keyLen := uint16(n.data[start])<<8 | uint16(n.data[start+1])
	valLen := uint16(n.data[start+2])<<8 | uint16(n.data[start+3])
	valStart := start + 4 + keyLen
	valEnd := valStart + valLen
	if int(valEnd) > len(n.data) {
		return nil
	}
	return n.data[valStart:valEnd]
}

// removeKV removes the entry at index pos.
func (n *Node) removeKV(pos int) {
	if pos < 0 || pos >= int(n.nkeys) {
		return
	}
	start := n.offsets[pos]
	if int(start)+4 > len(n.data) {
		return
	}
	keyLen := uint16(n.data[start])<<8 | uint16(n.data[start+1])
	valLen := uint16(n.data[start+2])<<8 | uint16(n.data[start+3])
	entrySize := int(4 + keyLen + valLen)
	end := start + uint16(entrySize)

	// Remove bytes from data slice
	n.data = append(n.data[:start], n.data[end:]...)

	// Remove offset from slice
	n.offsets = append(n.offsets[:pos], n.offsets[pos+1:]...)

	// Decrement following offsets
	for i := pos; i < len(n.offsets); i++ {
		n.offsets[i] -= uint16(entrySize)
	}

	n.nkeys--
}

// children returns the child nodes.
func (n *Node) children() []*Node {
	children := make([]*Node, len(n.pointers))
	for i := range n.pointers {
		children[i] = n.getChild(i)
	}
	return children
}