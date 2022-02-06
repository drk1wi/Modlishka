// Copyright 2020 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file at https://github.com/tidwall/btree/LICENSE

///////////////////////////////////////////////////////////////////////////////
// BEGIN PARAMS
///////////////////////////////////////////////////////////////////////////////

package btree

import "sync"

// degree is the B-Tree degree, which is equal to maximum number of children
// pre node times two.
// The default is 128, which means each node can have 255 items and 256 child
// nodes.
const degree = 128

// kind is the item type.
// It's important to use the equal symbol, which tells Go to create an alias of
// the type, rather than creating an entirely new type.
type kind = interface{}

// contextKind is the kind of context that can be passed to NewOptions and the
// less function
type contextKind = interface{}

// less returns true if A is less than B.
// The value of context will be whatever was passed to NewOptions through the
// Options.Context field, otherwise nil if the field was not set.
func less(a, b kind, context contextKind) bool {
	return context.(func(a, b contextKind) bool)(a, b)
}

// BTree aliases
// These are aliases to the local bTree types and functions, which are exported
// to allow for public use at a package level.
// Rename them if desired, or comment them out to make the library private.
type BTree = bTree
type Options = bOptions
type PathHint = bPathHint
type Iter = bIter

func New(less func(a, b kind) bool) *bTree { return bNew() }
func NewOptions(opts bOptions) *bTree      { return bNewOptions(opts) }

// The functions below, which begin with "test*", are required by the
// btree_test.go file. If you choose not use include the btree_test.go file in
// your project then these functions may be omitted.

// testCustomSeed can be used to generate a custom random seed for testing.
// Returning false will use time.Now().UnixNano()
func testCustomSeed() (seed int64, ok bool) {
	return 0, false
}

// testMakeItem must return a valid item for testing.
// It's required that the returned item maintains equal order as the
// provided int, such that:
//    testMakeItem(0) < testMakeItem(1) < testMakeItem(2) < testMakeItem(10)
func testMakeItem(x int) (item kind) {
	return x
}

// testNewBTree must return an operational btree for testing.
func testNewBTree() *bTree {
	return bNewOptions(bOptions{
		Context: func(a, b contextKind) bool {
			if a == nil {
				return b != nil
			} else if b == nil {
				return false
			}
			return a.(int) < b.(int)
		},
	})
}

///////////////////////////////////////////////////////////////////////////////
// END PARAMS
///////////////////////////////////////////////////////////////////////////////

// Do not edit code below this line.

const maxItems = degree*2 - 1 // max items per node. max children is +1
const minItems = maxItems / 2

type bTree struct {
	mu    *sync.RWMutex
	cow   *cow
	root  *node
	count int
	ctx   contextKind
	locks bool
	empty kind
}

type node struct {
	cow      *cow
	count    int
	items    []kind
	children *[]*node
}

type cow struct {
	_ int // cannot be an empty struct
}

func (tr *bTree) newNode(leaf bool) *node {
	n := &node{cow: tr.cow}
	if !leaf {
		n.children = new([]*node)
	}
	return n
}

// leaf returns true if the node is a leaf.
func (n *node) leaf() bool {
	return n.children == nil
}

// PathHint is a utility type used with the *Hint() functions. Hints provide
// faster operations for clustered keys.
type bPathHint struct {
	used [8]bool
	path [8]uint8
}

type bOptions struct {
	NoLocks bool
	Context contextKind
}

// New returns a new BTree
func bNew() *bTree {
	return bNewOptions(bOptions{})
}

func bNewOptions(opts bOptions) *bTree {
	tr := new(bTree)
	tr.cow = new(cow)
	tr.mu = new(sync.RWMutex)
	tr.ctx = opts.Context
	tr.locks = !opts.NoLocks
	return tr
}

// Less is a convenience function that performs a comparison of two items
// using the same "less" function provided to New.
func (tr *bTree) Less(a, b kind) bool {
	return less(a, b, tr.ctx)
}

func (tr *bTree) find(n *node, key kind,
	hint *bPathHint, depth int,
) (index int, found bool) {
	if hint == nil {
		// fast path for no hinting
		low := 0
		high := len(n.items)
		for low < high {
			mid := (low + high) / 2
			if !tr.Less(key, n.items[mid]) {
				low = mid + 1
			} else {
				high = mid
			}
		}
		if low > 0 && !tr.Less(n.items[low-1], key) {
			return low - 1, true
		}
		return low, false
	}

	// Try using hint.
	// Best case finds the exact match, updates the hint and returns.
	// Worst case, updates the low and high bounds to binary search between.
	low := 0
	high := len(n.items) - 1
	if depth < 8 && hint.used[depth] {
		index = int(hint.path[depth])
		if index >= len(n.items) {
			// tail item
			if tr.Less(n.items[len(n.items)-1], key) {
				index = len(n.items)
				goto path_match
			}
			index = len(n.items) - 1
		}
		if tr.Less(key, n.items[index]) {
			if index == 0 || tr.Less(n.items[index-1], key) {
				goto path_match
			}
			high = index - 1
		} else if tr.Less(n.items[index], key) {
			low = index + 1
		} else {
			found = true
			goto path_match
		}
	}

	// Do a binary search between low and high
	// keep on going until low > high, where the guarantee on low is that
	// key >= items[low - 1]
	for low <= high {
		mid := low + ((high+1)-low)/2
		// if key >= n.items[mid], low = mid + 1
		// which implies that key >= everything below low
		if !tr.Less(key, n.items[mid]) {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	// if low > 0, n.items[low - 1] >= key,
	// we have from before that key >= n.items[low - 1]
	// therefore key = n.items[low - 1],
	// and we have found the entry for key.
	// Otherwise we must keep searching for the key in index `low`.
	if low > 0 && !tr.Less(n.items[low-1], key) {
		index = low - 1
		found = true
	} else {
		index = low
		found = false
	}

path_match:
	if depth < 8 {
		hint.used[depth] = true
		var pathIndex uint8
		if n.leaf() && found {
			pathIndex = uint8(index + 1)
		} else {
			pathIndex = uint8(index)
		}
		if pathIndex != hint.path[depth] {
			hint.path[depth] = pathIndex
			for i := depth + 1; i < 8; i++ {
				hint.used[i] = false
			}
		}
	}
	return index, found
}

// SetHint sets or replace a value for a key using a path hint
func (tr *bTree) SetHint(item kind, hint *bPathHint) (prev kind, replaced bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	return tr.setHint(item, hint)
}

func (tr *bTree) setHint(item kind, hint *bPathHint) (prev kind, replaced bool) {
	if tr.root == nil {
		tr.root = tr.newNode(true)
		tr.root.items = append([]kind{}, item)
		tr.root.count = 1
		tr.count = 1
		return tr.empty, false
	}
	prev, replaced, split := tr.nodeSet(&tr.root, item, hint, 0)
	if split {
		left := tr.cowLoad(&tr.root)
		right, median := tr.nodeSplit(left)
		tr.root = tr.newNode(false)
		*tr.root.children = make([]*node, 0, maxItems+1)
		*tr.root.children = append([]*node{}, left, right)
		tr.root.items = append([]kind{}, median)
		tr.root.updateCount()
		return tr.setHint(item, hint)
	}
	if replaced {
		return prev, true
	}
	tr.count++
	return tr.empty, false
}

// Set or replace a value for a key
func (tr *bTree) Set(item kind) (kind, bool) {
	return tr.SetHint(item, nil)
}

func (tr *bTree) nodeSplit(n *node) (right *node, median kind) {
	i := maxItems / 2
	median = n.items[i]

	// left node
	left := tr.newNode(n.leaf())
	left.items = make([]kind, len(n.items[:i]), maxItems/2)
	copy(left.items, n.items[:i])
	if !n.leaf() {
		*left.children = make([]*node, len((*n.children)[:i+1]), maxItems+1)
		copy(*left.children, (*n.children)[:i+1])
	}
	left.updateCount()

	// right node
	right = tr.newNode(n.leaf())
	right.items = make([]kind, len(n.items[i+1:]), maxItems/2)
	copy(right.items, n.items[i+1:])
	if !n.leaf() {
		*right.children = make([]*node, len((*n.children)[i+1:]), maxItems+1)
		copy(*right.children, (*n.children)[i+1:])
	}
	right.updateCount()

	*n = *left
	return right, median
}

func (n *node) updateCount() {
	n.count = len(n.items)
	if !n.leaf() {
		for i := 0; i < len(*n.children); i++ {
			n.count += (*n.children)[i].count
		}
	}
}

// This operation should not be inlined because it's expensive and rarely
// called outside of heavy copy-on-write situations. Marking it "noinline"
// allows for the parent cowLoad to be inlined.
// go:noinline
func (tr *bTree) copy(n *node) *node {
	n2 := new(node)
	n2.cow = tr.cow
	n2.count = n.count
	n2.items = make([]kind, len(n.items), cap(n.items))
	copy(n2.items, n.items)
	if !n.leaf() {
		n2.children = new([]*node)
		*n2.children = make([]*node, len(*n.children), maxItems+1)
		copy(*n2.children, *n.children)
	}
	return n2
}

// cowLoad loads the provided node and, if needed, performs a copy-on-write.
func (tr *bTree) cowLoad(cn **node) *node {
	if (*cn).cow != tr.cow {
		*cn = tr.copy(*cn)
	}
	return *cn
}

func (tr *bTree) nodeSet(cn **node, item kind,
	hint *bPathHint, depth int,
) (prev kind, replaced bool, split bool) {
	n := tr.cowLoad(cn)
	i, found := tr.find(n, item, hint, depth)
	if found {
		prev = n.items[i]
		n.items[i] = item
		return prev, true, false
	}
	if n.leaf() {
		if len(n.items) == maxItems {
			return tr.empty, false, true
		}
		n.items = append(n.items, tr.empty)
		copy(n.items[i+1:], n.items[i:])
		n.items[i] = item
		n.count++
		return tr.empty, false, false
	}
	prev, replaced, split = tr.nodeSet(&(*n.children)[i], item, hint, depth+1)
	if split {
		if len(n.items) == maxItems {
			return tr.empty, false, true
		}
		right, median := tr.nodeSplit((*n.children)[i])
		*n.children = append(*n.children, nil)
		copy((*n.children)[i+1:], (*n.children)[i:])
		(*n.children)[i+1] = right
		n.items = append(n.items, tr.empty)
		copy(n.items[i+1:], n.items[i:])
		n.items[i] = median
		return tr.nodeSet(&n, item, hint, depth)
	}
	if !replaced {
		n.count++
	}
	return prev, replaced, false
}

func (tr *bTree) Scan(iter func(item kind) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return
	}
	tr.root.scan(iter)
}

func (n *node) scan(iter func(item kind) bool) bool {
	if n.leaf() {
		for i := 0; i < len(n.items); i++ {
			if !iter(n.items[i]) {
				return false
			}
		}
		return true
	}
	for i := 0; i < len(n.items); i++ {
		if !(*n.children)[i].scan(iter) {
			return false
		}
		if !iter(n.items[i]) {
			return false
		}
	}
	return (*n.children)[len(*n.children)-1].scan(iter)
}

// Get a value for key
func (tr *bTree) Get(key kind) (kind, bool) {
	return tr.GetHint(key, nil)
}

// GetHint gets a value for key using a path hint
func (tr *bTree) GetHint(key kind, hint *bPathHint) (kind, bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.root
	depth := 0
	for {
		i, found := tr.find(n, key, hint, depth)
		if found {
			return n.items[i], true
		}
		if n.children == nil {
			return tr.empty, false
		}
		n = (*n.children)[i]
		depth++
	}
}

// Len returns the number of items in the tree
func (tr *bTree) Len() int {
	return tr.count
}

// Delete a value for a key
func (tr *bTree) Delete(key kind) (kind, bool) {
	return tr.DeleteHint(key, nil)
}

// DeleteHint deletes a value for a key using a path hint
func (tr *bTree) DeleteHint(key kind, hint *bPathHint) (kind, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	return tr.deleteHint(key, hint)
}

func (tr *bTree) deleteHint(key kind, hint *bPathHint) (kind, bool) {
	if tr.root == nil {
		return tr.empty, false
	}
	prev, deleted := tr.delete(&tr.root, false, key, hint, 0)
	if !deleted {
		return tr.empty, false
	}
	if len(tr.root.items) == 0 && !tr.root.leaf() {
		tr.root = (*tr.root.children)[0]
	}
	tr.count--
	if tr.count == 0 {
		tr.root = nil
	}
	return prev, true
}

func (tr *bTree) delete(cn **node, max bool, key kind,
	hint *bPathHint, depth int,
) (kind, bool) {
	n := tr.cowLoad(cn)
	var i int
	var found bool
	if max {
		i, found = len(n.items)-1, true
	} else {
		i, found = tr.find(n, key, hint, depth)
	}
	if n.leaf() {
		if found {
			// found the items at the leaf, remove it and return.
			prev := n.items[i]
			copy(n.items[i:], n.items[i+1:])
			n.items[len(n.items)-1] = tr.empty
			n.items = n.items[:len(n.items)-1]
			n.count--
			return prev, true
		}
		return tr.empty, false
	}

	var prev kind
	var deleted bool
	if found {
		if max {
			i++
			prev, deleted = tr.delete(&(*n.children)[i], true, tr.empty, nil, 0)
		} else {
			prev = n.items[i]
			maxItem, _ := tr.delete(&(*n.children)[i], true, tr.empty, nil, 0)
			deleted = true
			n.items[i] = maxItem
		}
	} else {
		prev, deleted = tr.delete(&(*n.children)[i], max, key, hint, depth+1)
	}
	if !deleted {
		return tr.empty, false
	}
	n.count--
	if len((*n.children)[i].items) < minItems {
		tr.nodeRebalance(n, i)
	}
	return prev, true

}

// nodeRebalance rebalances the child nodes following a delete operation.
// Provide the index of the child node with the number of items that fell
// below minItems.
func (tr *bTree) nodeRebalance(n *node, i int) {
	if i == len(n.items) {
		i--
	}

	// ensure copy-on-write
	left := tr.cowLoad(&(*n.children)[i])
	right := tr.cowLoad(&(*n.children)[i+1])

	if len(left.items)+len(right.items) < maxItems {
		// Merges the left and right children nodes together as a single node
		// that includes (left,item,right), and places the contents into the
		// existing left node. Delete the right node altogether and move the
		// following items and child nodes to the left by one slot.

		// merge (left,item,right)
		left.items = append(left.items, n.items[i])
		left.items = append(left.items, right.items...)
		if !left.leaf() {
			*left.children = append(*left.children, *right.children...)
		}
		left.count += right.count + 1

		// move the items over one slot
		copy(n.items[i:], n.items[i+1:])
		n.items[len(n.items)-1] = tr.empty
		n.items = n.items[:len(n.items)-1]

		// move the children over one slot
		copy((*n.children)[i+1:], (*n.children)[i+2:])
		(*n.children)[len(*n.children)-1] = nil
		(*n.children) = (*n.children)[:len(*n.children)-1]
	} else if len(left.items) > len(right.items) {
		// move left -> right over one slot

		// Move the item of the parent node at index into the right-node first
		// slot, and move the left-node last item into the previously moved
		// parent item slot.
		right.items = append(right.items, tr.empty)
		copy(right.items[1:], right.items)
		right.items[0] = n.items[i]
		right.count++
		n.items[i] = left.items[len(left.items)-1]
		left.items[len(left.items)-1] = tr.empty
		left.items = left.items[:len(left.items)-1]
		left.count--

		if !left.leaf() {
			// move the left-node last child into the right-node first slot
			*right.children = append(*right.children, nil)
			copy((*right.children)[1:], *right.children)
			(*right.children)[0] = (*left.children)[len(*left.children)-1]
			(*left.children)[len(*left.children)-1] = nil
			(*left.children) = (*left.children)[:len(*left.children)-1]
			left.count -= (*right.children)[0].count
			right.count += (*right.children)[0].count
		}
	} else {
		// move left <- right over one slot

		// Same as above but the other direction
		left.items = append(left.items, n.items[i])
		left.count++
		n.items[i] = right.items[0]
		copy(right.items, right.items[1:])
		right.items[len(right.items)-1] = tr.empty
		right.items = right.items[:len(right.items)-1]
		right.count--

		if !left.leaf() {
			*left.children = append(*left.children, (*right.children)[0])
			copy(*right.children, (*right.children)[1:])
			(*right.children)[len(*right.children)-1] = nil
			*right.children = (*right.children)[:len(*right.children)-1]
			left.count += (*left.children)[len(*left.children)-1].count
			right.count -= (*left.children)[len(*left.children)-1].count
		}
	}
}

// Ascend the tree within the range [pivot, last]
// Pass nil for pivot to scan all item in ascending order
// Return false to stop iterating
func (tr *bTree) Ascend(pivot kind, iter func(item kind) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return
	}
	tr.ascend(tr.root, pivot, nil, 0, iter)
}

// The return value of this function determines whether we should keep iterating
// upon this functions return.
func (tr *bTree) ascend(n *node, pivot kind,
	hint *bPathHint, depth int, iter func(item kind) bool,
) bool {
	i, found := tr.find(n, pivot, hint, depth)
	if !found {
		if !n.leaf() {
			if !tr.ascend((*n.children)[i], pivot, hint, depth+1, iter) {
				return false
			}
		}
	}
	// We are either in the case that
	// - node is found, we should iterate through it starting at `i`,
	//   the index it was located at.
	// - node is not found, and TODO: fill in.
	for ; i < len(n.items); i++ {
		if !iter(n.items[i]) {
			return false
		}
		if !n.leaf() {
			if !(*n.children)[i+1].scan(iter) {
				return false
			}
		}
	}
	return true
}

func (tr *bTree) Reverse(iter func(item kind) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return
	}
	tr.root.reverse(iter)
}

func (n *node) reverse(iter func(item kind) bool) bool {
	if n.leaf() {
		for i := len(n.items) - 1; i >= 0; i-- {
			if !iter(n.items[i]) {
				return false
			}
		}
		return true
	}
	if !(*n.children)[len(*n.children)-1].reverse(iter) {
		return false
	}
	for i := len(n.items) - 1; i >= 0; i-- {
		if !iter(n.items[i]) {
			return false
		}
		if !(*n.children)[i].reverse(iter) {
			return false
		}
	}
	return true
}

// Descend the tree within the range [pivot, first]
// Pass nil for pivot to scan all item in descending order
// Return false to stop iterating
func (tr *bTree) Descend(pivot kind, iter func(item kind) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return
	}
	tr.descend(tr.root, pivot, nil, 0, iter)
}

func (tr *bTree) descend(n *node, pivot kind,
	hint *bPathHint, depth int, iter func(item kind) bool,
) bool {
	i, found := tr.find(n, pivot, hint, depth)
	if !found {
		if !n.leaf() {
			if !tr.descend((*n.children)[i], pivot, hint, depth+1, iter) {
				return false
			}
		}
		i--
	}
	for ; i >= 0; i-- {
		if !iter(n.items[i]) {
			return false
		}
		if !n.leaf() {
			if !(*n.children)[i].reverse(iter) {
				return false
			}
		}
	}
	return true
}

// Load is for bulk loading pre-sorted items
func (tr *bTree) Load(item kind) (kind, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	if tr.root == nil {
		return tr.setHint(item, nil)
	}
	n := tr.cowLoad(&tr.root)
	for {
		n.count++ // optimistically update counts
		if n.leaf() {
			if len(n.items) < maxItems {
				if tr.Less(n.items[len(n.items)-1], item) {
					n.items = append(n.items, item)
					tr.count++
					return tr.empty, false
				}
			}
			break
		}
		n = tr.cowLoad(&(*n.children)[len(*n.children)-1])
	}
	// revert the counts
	n = tr.root
	for {
		n.count--
		if n.leaf() {
			break
		}
		n = (*n.children)[len(*n.children)-1]
	}
	return tr.setHint(item, nil)
}

// Min returns the minimum item in tree.
// Returns nil if the tree has no items.
func (tr *bTree) Min() (kind, bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.root
	for {
		if n.leaf() {
			return n.items[0], true
		}
		n = (*n.children)[0]
	}
}

// Max returns the maximum item in tree.
// Returns nil if the tree has no items.
func (tr *bTree) Max() (kind, bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.root
	for {
		if n.leaf() {
			return n.items[len(n.items)-1], true
		}
		n = (*n.children)[len(*n.children)-1]
	}
}

// PopMin removes the minimum item in tree and returns it.
// Returns nil if the tree has no items.
func (tr *bTree) PopMin() (kind, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.cowLoad(&tr.root)
	var item kind
	for {
		n.count-- // optimistically update counts
		if n.leaf() {
			item = n.items[0]
			if len(n.items) == minItems {
				break
			}
			copy(n.items[:], n.items[1:])
			n.items[len(n.items)-1] = tr.empty
			n.items = n.items[:len(n.items)-1]
			tr.count--
			if tr.count == 0 {
				tr.root = nil
			}
			return item, true
		}
		n = tr.cowLoad(&(*n.children)[0])
	}
	// revert the counts
	n = tr.root
	for {
		n.count++
		if n.leaf() {
			break
		}
		n = (*n.children)[0]
	}
	return tr.deleteHint(item, nil)
}

// PopMax removes the minimum item in tree and returns it.
// Returns nil if the tree has no items.
func (tr *bTree) PopMax() (kind, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.cowLoad(&tr.root)
	var item kind
	for {
		n.count-- // optimistically update counts
		if n.leaf() {
			item = n.items[len(n.items)-1]
			if len(n.items) == minItems {
				break
			}
			n.items[len(n.items)-1] = tr.empty
			n.items = n.items[:len(n.items)-1]
			tr.count--
			if tr.count == 0 {
				tr.root = nil
			}
			return item, true
		}
		n = tr.cowLoad(&(*n.children)[len(*n.children)-1])
	}
	// revert the counts
	n = tr.root
	for {
		n.count++
		if n.leaf() {
			break
		}
		n = (*n.children)[len(*n.children)-1]
	}
	return tr.deleteHint(item, nil)
}

// GetAt returns the value at index.
// Return nil if the tree is empty or the index is out of bounds.
func (tr *bTree) GetAt(index int) (kind, bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil || index < 0 || index >= tr.count {
		return tr.empty, false
	}
	n := tr.root
	for {
		if n.leaf() {
			return n.items[index], true
		}
		i := 0
		for ; i < len(n.items); i++ {
			if index < (*n.children)[i].count {
				break
			} else if index == (*n.children)[i].count {
				return n.items[i], true
			}
			index -= (*n.children)[i].count + 1
		}
		n = (*n.children)[i]
	}
}

// DeleteAt deletes the item at index.
// Return nil if the tree is empty or the index is out of bounds.
func (tr *bTree) DeleteAt(index int) (kind, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	if tr.root == nil || index < 0 || index >= tr.count {
		return tr.empty, false
	}
	var pathbuf [8]uint8 // track the path
	path := pathbuf[:0]
	var item kind
	n := tr.cowLoad(&tr.root)
outer:
	for {
		n.count-- // optimistically update counts
		if n.leaf() {
			// the index is the item position
			item = n.items[index]
			if len(n.items) == minItems {
				path = append(path, uint8(index))
				break outer
			}
			copy(n.items[index:], n.items[index+1:])
			n.items[len(n.items)-1] = tr.empty
			n.items = n.items[:len(n.items)-1]
			tr.count--
			if tr.count == 0 {
				tr.root = nil
			}
			return item, true
		}
		i := 0
		for ; i < len(n.items); i++ {
			if index < (*n.children)[i].count {
				break
			} else if index == (*n.children)[i].count {
				item = n.items[i]
				path = append(path, uint8(i))
				break outer
			}
			index -= (*n.children)[i].count + 1
		}
		path = append(path, uint8(i))
		n = tr.cowLoad(&(*n.children)[i])
	}
	// revert the counts
	var hint bPathHint
	n = tr.root
	for i := 0; i < len(path); i++ {
		if i < len(hint.path) {
			hint.path[i] = uint8(path[i])
			hint.used[i] = true
		}
		n.count++
		if !n.leaf() {
			n = (*n.children)[uint8(path[i])]
		}
	}
	return tr.deleteHint(item, &hint)
}

// Height returns the height of the tree.
// Returns zero if tree has no items.
func (tr *bTree) Height() int {
	if tr.rlock() {
		defer tr.runlock()
	}
	var height int
	if tr.root != nil {
		n := tr.root
		for {
			height++
			if n.leaf() {
				break
			}
			n = (*n.children)[0]
		}
	}
	return height
}

// Walk iterates over all items in tree, in order.
// The items param will contain one or more items.
func (tr *bTree) Walk(iter func(item []kind) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root != nil {
		tr.root.walk(iter)
	}
}

func (n *node) walk(iter func(item []kind) bool) bool {
	if n.leaf() {
		if !iter(n.items) {
			return false
		}
	} else {
		for i := 0; i < len(n.items); i++ {
			(*n.children)[i].walk(iter)
			if !iter(n.items[i : i+1]) {
				return false
			}
		}
		(*n.children)[len(n.items)].walk(iter)
	}
	return true
}

// Copy the tree. This is a copy-on-write operation and is very fast because
// it only performs a shadowed copy.
func (tr *bTree) Copy() *bTree {
	if tr.lock() {
		defer tr.unlock()
	}
	tr.cow = new(cow)
	tr2 := new(bTree)
	*tr2 = *tr
	tr2.mu = new(sync.RWMutex)
	tr2.cow = new(cow)
	return tr2
}

func (tr *bTree) lock() bool {
	if tr.locks {
		tr.mu.Lock()
	}
	return tr.locks
}

func (tr *bTree) unlock() {
	tr.mu.Unlock()
}

func (tr *bTree) rlock() bool {
	if tr.locks {
		tr.mu.RLock()
	}
	return tr.locks
}

func (tr *bTree) runlock() {
	tr.mu.RUnlock()
}

// Iter represents an iterator
type bIter struct {
	tr      *bTree
	locked  bool
	seeked  bool
	atstart bool
	atend   bool
	stack   []iterStackItem
	item    kind
}

type iterStackItem struct {
	n *node
	i int
}

// Iter returns a read-only iterator.
// The Release method must be called finished with iterator.
func (tr *bTree) Iter() bIter {
	var iter bIter
	iter.tr = tr
	iter.locked = tr.rlock()
	return iter
}

// Seek to item greater-or-equal-to key.
// Returns false if there was no item found.
func (iter *bIter) Seek(key kind) bool {
	if iter.tr == nil {
		return false
	}
	iter.seeked = true
	iter.stack = iter.stack[:0]
	if iter.tr.root == nil {
		return false
	}
	n := iter.tr.root
	for {
		i, found := iter.tr.find(n, key, nil, 0)
		iter.stack = append(iter.stack, iterStackItem{n, i})
		if found {
			return true
		}
		if n.leaf() {
			if i == len(n.items) {
				iter.stack = iter.stack[:0]
				return false
			}
			return true
		}
		n = (*n.children)[i]
	}
}

// First moves iterator to first item in tree.
// Returns false if the tree is empty.
func (iter *bIter) First() bool {
	if iter.tr == nil {
		return false
	}
	iter.atend = false
	iter.atstart = false
	iter.seeked = true
	iter.stack = iter.stack[:0]
	if iter.tr.root == nil {
		return false
	}
	n := iter.tr.root
	for {
		iter.stack = append(iter.stack, iterStackItem{n, 0})
		if n.leaf() {
			break
		}
		n = (*n.children)[0]
	}
	s := &iter.stack[len(iter.stack)-1]
	iter.item = s.n.items[s.i]
	return true
}

// Last moves iterator to last item in tree.
// Returns false if the tree is empty.
func (iter *bIter) Last() bool {
	if iter.tr == nil {
		return false
	}
	iter.seeked = true
	iter.stack = iter.stack[:0]
	if iter.tr.root == nil {
		return false
	}
	n := iter.tr.root
	for {
		iter.stack = append(iter.stack, iterStackItem{n, len(n.items)})
		if n.leaf() {
			iter.stack[len(iter.stack)-1].i--
			break
		}
		n = (*n.children)[len(n.items)]
	}
	s := &iter.stack[len(iter.stack)-1]
	iter.item = s.n.items[s.i]
	return true
}

// First moves iterator to first item in tree.
// Returns false if the tree is empty.
func (iter *bIter) Release() {
	if iter.tr == nil {
		return
	}
	if iter.locked {
		iter.tr.runlock()
		iter.locked = false
	}
	iter.stack = nil
	iter.tr = nil
}

// Next moves iterator to the next item in iterator.
// Returns false if the tree is empty or the iterator is at the end of
// the tree.
func (iter *bIter) Next() bool {
	if iter.tr == nil {
		return false
	}
	if !iter.seeked {
		return iter.First()
	}
	if len(iter.stack) == 0 {
		if iter.atstart {
			return iter.First() && iter.Next()
		}
		return false
	}
	s := &iter.stack[len(iter.stack)-1]
	s.i++
	if s.n.leaf() {
		if s.i == len(s.n.items) {
			for {
				iter.stack = iter.stack[:len(iter.stack)-1]
				if len(iter.stack) == 0 {
					iter.atend = true
					return false
				}
				s = &iter.stack[len(iter.stack)-1]
				if s.i < len(s.n.items) {
					break
				}
			}
		}
	} else {
		n := (*s.n.children)[s.i]
		for {
			iter.stack = append(iter.stack, iterStackItem{n, 0})
			if n.leaf() {
				break
			}
			n = (*n.children)[0]
		}
	}
	s = &iter.stack[len(iter.stack)-1]
	iter.item = s.n.items[s.i]
	return true
}

// Prev moves iterator to the previous item in iterator.
// Returns false if the tree is empty or the iterator is at the beginning of
// the tree.
func (iter *bIter) Prev() bool {
	if iter.tr == nil {
		return false
	}
	if !iter.seeked {
		return false
	}
	if len(iter.stack) == 0 {
		if iter.atend {
			return iter.Last() && iter.Prev()
		}
		return false
	}
	s := &iter.stack[len(iter.stack)-1]
	if s.n.leaf() {
		s.i--
		if s.i == -1 {
			for {
				iter.stack = iter.stack[:len(iter.stack)-1]
				if len(iter.stack) == 0 {
					iter.atstart = true
					return false
				}
				s = &iter.stack[len(iter.stack)-1]
				s.i--
				if s.i > -1 {
					break
				}
			}
		}
	} else {
		n := (*s.n.children)[s.i]
		for {
			iter.stack = append(iter.stack, iterStackItem{n, len(n.items)})
			if n.leaf() {
				iter.stack[len(iter.stack)-1].i--
				break
			}
			n = (*n.children)[len(n.items)]
		}
	}
	s = &iter.stack[len(iter.stack)-1]
	iter.item = s.n.items[s.i]
	return true
}

// Item returns the current iterator item.
func (iter *bIter) Item() kind {
	return iter.item
}
