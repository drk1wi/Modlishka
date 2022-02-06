// Copyright 2020 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package btree

import btree "github.com/tidwall/btree/internal"

type BTree struct {
	base *btree.BTree
}

// PathHint is a utility type used with the *Hint() functions. Hints provide
// faster operations for clustered keys.
type PathHint = btree.PathHint

// New returns a new BTree
func New(less func(a, b interface{}) bool) *BTree {
	if less == nil {
		panic("nil less")
	}
	return &BTree{
		base: btree.NewOptions(btree.Options{
			Context: less,
		}),
	}
}

// NewNonConcurrent returns a new BTree which is not safe for concurrent
// write operations by multiple goroutines.
//
// This is useful for when you do not need the BTree to manage the locking,
// but would rather do it yourself.
func NewNonConcurrent(less func(a, b interface{}) bool) *BTree {
	if less == nil {
		panic("nil less")
	}
	return &BTree{
		base: btree.NewOptions(btree.Options{
			Context: less,
			NoLocks: true,
		}),
	}
}

// Less is a convenience function that performs a comparison of two items
// using the same "less" function provided to New.
func (tr *BTree) Less(a, b interface{}) bool {
	return tr.base.Less(a, b)
}

// Set or replace a value for a key
func (tr *BTree) Set(item interface{}) interface{} {
	return tr.SetHint(item, nil)
}

// SetHint sets or replace a value for a key using a path hint
func (tr *BTree) SetHint(item interface{}, hint *PathHint) (prev interface{}) {
	if item == nil {
		panic("nil item")
	}
	v, ok := tr.base.SetHint(item, hint)
	if !ok {
		return nil
	}
	return v
}

// Get a value for key
func (tr *BTree) Get(key interface{}) interface{} {
	return tr.GetHint(key, nil)
}

// GetHint gets a value for key using a path hint
func (tr *BTree) GetHint(key interface{}, hint *PathHint) interface{} {
	if key == nil {
		return nil
	}
	v, ok := tr.base.GetHint(key, hint)
	if !ok {
		return nil
	}
	return v
}

// Len returns the number of items in the tree
func (tr *BTree) Len() int {
	return tr.base.Len()
}

// Delete a value for a key
func (tr *BTree) Delete(key interface{}) interface{} {
	return tr.DeleteHint(key, nil)
}

// DeleteHint deletes a value for a key using a path hint
func (tr *BTree) DeleteHint(key interface{}, hint *PathHint) interface{} {
	if key == nil {
		return nil
	}
	v, ok := tr.base.DeleteHint(key, nil)
	if !ok {
		return nil
	}
	return v
}

// Ascend the tree within the range [pivot, last]
// Pass nil for pivot to scan all item in ascending order
// Return false to stop iterating
func (tr *BTree) Ascend(pivot interface{}, iter func(item interface{}) bool) {
	if pivot == nil {
		tr.base.Scan(iter)
	} else {
		tr.base.Ascend(pivot, iter)
	}
}

// Descend the tree within the range [pivot, first]
// Pass nil for pivot to scan all item in descending order
// Return false to stop iterating
func (tr *BTree) Descend(pivot interface{}, iter func(item interface{}) bool) {
	if pivot == nil {
		tr.base.Reverse(iter)
	} else {
		tr.base.Descend(pivot, iter)
	}
}

// Load is for bulk loading pre-sorted items
func (tr *BTree) Load(item interface{}) interface{} {
	if item == nil {
		panic("nil item")
	}
	v, ok := tr.base.Load(item)
	if !ok {
		return nil
	}
	return v
}

// Min returns the minimum item in tree.
// Returns nil if the tree has no items.
func (tr *BTree) Min() interface{} {
	v, ok := tr.base.Min()
	if !ok {
		return nil
	}
	return v
}

// Max returns the maximum item in tree.
// Returns nil if the tree has no items.
func (tr *BTree) Max() interface{} {
	v, ok := tr.base.Max()
	if !ok {
		return nil
	}
	return v
}

// PopMin removes the minimum item in tree and returns it.
// Returns nil if the tree has no items.
func (tr *BTree) PopMin() interface{} {
	v, ok := tr.base.PopMin()
	if !ok {
		return nil
	}
	return v
}

// PopMax removes the minimum item in tree and returns it.
// Returns nil if the tree has no items.
func (tr *BTree) PopMax() interface{} {
	v, ok := tr.base.PopMax()
	if !ok {
		return nil
	}
	return v
}

// GetAt returns the value at index.
// Return nil if the tree is empty or the index is out of bounds.
func (tr *BTree) GetAt(index int) interface{} {
	v, ok := tr.base.GetAt(index)
	if !ok {
		return nil
	}
	return v
}

// DeleteAt deletes the item at index.
// Return nil if the tree is empty or the index is out of bounds.
func (tr *BTree) DeleteAt(index int) interface{} {
	v, ok := tr.base.DeleteAt(index)
	if !ok {
		return nil
	}
	return v
}

// Height returns the height of the tree.
// Returns zero if tree has no items.
func (tr *BTree) Height() int {
	return tr.base.Height()
}

// Walk iterates over all items in tree, in order.
// The items param will contain one or more items.
func (tr *BTree) Walk(iter func(items []interface{})) {
	tr.base.Walk(func(items []interface{}) bool {
		iter(items)
		return true
	})
}

// Copy the tree. This is a copy-on-write operation and is very fast because
// it only performs a shadowed copy.
func (tr *BTree) Copy() *BTree {
	return &BTree{base: tr.base.Copy()}
}

type Iter struct {
	base btree.Iter
}

// Iter returns a read-only iterator.
// The Release method must be called finished with iterator.
func (tr *BTree) Iter() Iter {
	return Iter{tr.base.Iter()}
}

// Seek to item greater-or-equal-to key.
// Returns false if there was no item found.
func (iter *Iter) Seek(key interface{}) bool {
	return iter.base.Seek(key)
}

// First moves iterator to first item in tree.
// Returns false if the tree is empty.
func (iter *Iter) First() bool {
	return iter.base.First()
}

// Last moves iterator to last item in tree.
// Returns false if the tree is empty.
func (iter *Iter) Last() bool {
	return iter.base.Last()
}

// First moves iterator to first item in tree.
// Returns false if the tree is empty.
func (iter *Iter) Release() {
	iter.base.Release()
}

// Next moves iterator to the next item in iterator.
// Returns false if the tree is empty or the iterator is at the end of
// the tree.
func (iter *Iter) Next() bool {
	return iter.base.Next()
}

// Prev moves iterator to the previous item in iterator.
// Returns false if the tree is empty or the iterator is at the beginning of
// the tree.
func (iter *Iter) Prev() bool {
	return iter.base.Prev()
}

// Item returns the current iterator item.
func (iter *Iter) Item() interface{} {
	return iter.base.Item()
}
