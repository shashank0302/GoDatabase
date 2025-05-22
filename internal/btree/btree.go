// Package btree provides a B+Tree implementation for the storage engine.
// The B+Tree is a balanced tree data structure that maintains sorted data
// and allows searches, sequential access, insertions, and deletions in O(log n) time.
package btree

import (
	"bytes"
	"errors"
)

// BTree represents the overall B+Tree data structure.
// A B+Tree is a self-balancing tree data structure that maintains sorted data
// and allows searches, sequential access, insertions, and deletions in logarithmic time.
type BTree struct {
	root *Node // The root node of the tree
	size int   // The number of keys in the tree
}

// NewBTree creates a new B+ tree with an empty leaf node as the root.
//
// Returns:
//   - A pointer to a new BTree instance
func NewBTree() *BTree {
	// Create a new leaf node as the root
	root := NewNode(BNODE_LEAF)
	return &BTree{
		root: root,
		size: 0,
	}
}

// Insert adds a key/value pair into the B+ tree.
// The method validates the inputs, finds the appropriate leaf node,
// inserts the key/value pair, and handles any necessary node splitting.
//
// Parameters:
//   - key: The key as a byte slice
//   - value: The value as a byte slice
//
// Returns:
//   - An error if the key is too large, value is too large, or key already exists
func (t *BTree) Insert(key, value []byte) error {
	// Validate input
	if len(key) > BTREE_MAX_KEY_SIZE {
		return errors.New("key too large")
	}
	if len(value) > BTREE_MAX_VAL_SIZE {
		return errors.New("value too large")
	}

	// Find the leaf node where the key should be inserted
	leaf := t.findLeaf(t.root, key)
	
	// Insert the key/value pair into the leaf
	if err := t.insertInLeaf(leaf, key, value); err != nil {
		return err
	}
	
	// If the leaf is now overfull, split it
	if leaf.IsFull() {
		newLeaf, promotedKey := leaf.Split()
		// Propagate the split upward
		t.insertInParent(leaf, promotedKey, newLeaf)
	}

	t.size++
	return nil
}

// findLeaf traverses the tree to find the leaf node where a key belongs.
// It performs a recursive search starting from the provided node.
//
// Parameters:
//   - n: The node to start the search from
//   - key: The key to find the leaf for
//
// Returns:
//   - A pointer to the leaf Node where key belongs
func (t *BTree) findLeaf(n *Node, key []byte) *Node {
	// If node is leaf, return it
	if n.typ == BNODE_LEAF {
		return n
	}
	
	// For internal node, choose the proper child pointer
	// by comparing the key with each key in the node
	for i, k := range n.keys() {
		if bytes.Compare(key, k) < 0 {
			// Key is smaller than the current node key,
			// so go down the left child pointer
			return t.findLeaf(n.getChild(i), key)
		}
	}
	// Otherwise, key is greater than all keys in n; use last child
	// This follows the B+Tree property where keys in a node divide
	// the key space for its children
	return t.findLeaf(n.getChild(len(n.keys())), key)
}

// insertInLeaf inserts a key/value pair into a leaf node in sorted order.
// It finds the correct position for the key and delegates the actual insertion
// to the node's insertKV method.
//
// Parameters:
//   - leaf: The leaf node to insert into
//   - key: The key to insert
//   - value: The value to insert
//
// Returns:
//   - An error if the key already exists
func (t *BTree) insertInLeaf(leaf *Node, key, value []byte) error {
	// Find insertion position
	pos := 0
	for i, k := range leaf.keys() {
		if bytes.Compare(key, k) == 0 {
			return errors.New("key already exists")
		}
		if bytes.Compare(key, k) < 0 {
			break
		}
		pos = i + 1
	}

	// Insert key and value
	leaf.insertKV(pos, key, value)
	return nil
}

// insertInParent handles the upward propagation after a node split.
// This is a key part of maintaining the B+Tree structure when a node becomes too large.
//
// Parameters:
//   - oldNode: The original node that was split
//   - key: The key that was promoted from the split
//   - newNode: The new node created from the split
func (t *BTree) insertInParent(oldNode *Node, key []byte, newNode *Node) {
	// If oldNode is root, create a new root
	if oldNode == t.root {
		newRoot := NewNode(BNODE_NODE)
		newRoot.insertKV(0, key, nil)
		newRoot.setChild(0, oldNode)
		newRoot.setChild(1, newNode)
		t.root = newRoot
		return
	}

	// Find the parent node
	parent := t.findParent(t.root, oldNode)
	if parent == nil {
		panic("parent not found")
	}

	// Insert key and newNode pointer into the parent
	pos := 0
	for i, k := range parent.keys() {
		if bytes.Compare(key, k) < 0 {
			break
		}
		pos = i + 1
	}
	parent.insertKV(pos, key, nil)
	parent.setChild(pos+1, newNode)

	// If parent overflows, split it recursively
	if parent.IsFull() {
		newParent, promotedKey := parent.Split()
		t.insertInParent(parent, promotedKey, newParent)
	}
}

// findParent finds the parent node of a given node by traversing the tree.
// This is used during insertInParent to locate where changes need to be made.
//
// Parameters:
//   - root: The node to start the search from
//   - target: The node whose parent we're looking for
//
// Returns:
//   - A pointer to the parent Node, or nil if not found
func (t *BTree) findParent(root, target *Node) *Node {
	if root == target {
		return nil
	}

	if root.typ == BNODE_LEAF {
		return nil
	}

	for i := 0; i < len(root.pointers); i++ {
		child := root.getChild(i)
		if child == target {
			return root
		}
		if found := t.findParent(child, target); found != nil {
			return found
		}
	}
	return nil
}

// Get retrieves a value for a given key from the B+Tree.
// It traverses to the correct leaf node and searches for the key.
//
// Parameters:
//   - key: The key to look up
//
// Returns:
//   - The value as a byte slice
//   - An error if the key is not found
func (t *BTree) Get(key []byte) ([]byte, error) {
	// Find the leaf node where the key should be
	leaf := t.findLeaf(t.root, key)
	
	// Search for the key in the leaf node
	for i, k := range leaf.keys() {
		if bytes.Compare(key, k) == 0 {
			return leaf.getValue(i), nil
		}
	}
	return nil, errors.New("key not found")
}

// Delete removes a key/value pair from the B+ tree.
// It finds the key, removes it, and handles any necessary tree rebalancing.
//
// Parameters:
//   - key: The key to delete
//
// Returns:
//   - An error if the key is not found
func (t *BTree) Delete(key []byte) error {
	// Find the leaf containing the key
	leaf := t.findLeaf(t.root, key)
	
	// Search for the key's position in the leaf
	pos := -1
	for i, k := range leaf.keys() {
		if bytes.Compare(key, k) == 0 {
			pos = i
			break
		}
	}
	if pos == -1 {
		return errors.New("key not found")
	}

	// Remove the key/value pair
	leaf.removeKV(pos)

	// If the leaf is now underfull, try to redistribute or merge
	if leaf.IsEmpty() && leaf != t.root {
		t.rebalance(leaf)
	}

	t.size--
	return nil
}

// rebalance handles underflow in a node by redistributing keys or merging nodes.
// This ensures the B+Tree remains balanced after deletions.
//
// Parameters:
//   - n: The node to rebalance
func (t *BTree) rebalance(n *Node) {
	parent := t.findParent(t.root, n)
	if parent == nil {
		return
	}

	// Find the position of n in parent's children
	pos := -1
	for i, child := range parent.children() {
		if child == n {
			pos = i
			break
		}
	}
	if pos == -1 {
		panic("node not found in parent")
	}

	// Try to redistribute with left sibling
	if pos > 0 {
		leftSibling := parent.getChild(pos - 1)
		if !leftSibling.IsFull() {
			t.redistribute(leftSibling, n, parent, pos-1)
			return
		}
	}

	// Try to redistribute with right sibling
	if pos < len(parent.children())-1 {
		rightSibling := parent.getChild(pos + 1)
		if !rightSibling.IsFull() {
			t.redistribute(n, rightSibling, parent, pos)
			return
		}
	}

	// If redistribution failed, merge
	if pos > 0 {
		leftSibling := parent.getChild(pos - 1)
		t.merge(leftSibling, n, parent, pos-1)
	} else {
		rightSibling := parent.getChild(pos + 1)
		t.merge(n, rightSibling, parent, pos)
	}
}

// redistribute moves keys between two nodes to balance them.
// This is a simplified implementation that needs to be expanded for a full B+Tree.
//
// Parameters:
//   - left: The left node
//   - right: The right node
//   - parent: The parent node
//   - pos: The position of the separator key in the parent
func (t *BTree) redistribute(left, right *Node, parent *Node, pos int) {
	// Implementation of redistribution logic
	// This is a simplified version - you'll need to implement the full logic
	// based on your specific requirements
}

// merge combines two nodes into one.
// This is a simplified implementation that needs to be expanded for a full B+Tree.
//
// Parameters:
//   - left: The left node
//   - right: The right node
//   - parent: The parent node
//   - pos: The position of the separator key in the parent
func (t *BTree) merge(left, right *Node, parent *Node, pos int) {
	// Implementation of merge logic
	// This is a simplified version - you'll need to implement the full logic
	// based on your specific requirements
}

// Size returns the number of keys in the tree.
//
// Returns:
//   - The size of the tree (number of key-value pairs)
func (t *BTree) Size() int {
	return t.size
}

// Height returns the height of the tree.
// The height is the number of levels in the tree.
//
// Returns:
//   - The height of the tree
func (t *BTree) Height() int {
	height := 0
	node := t.root
	for node.typ != BNODE_LEAF {
		height++
		node = node.getChild(0)
	}
	return height
}