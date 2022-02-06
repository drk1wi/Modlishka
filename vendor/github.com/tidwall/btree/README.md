# btree

[![GoDoc](https://godoc.org/github.com/tidwall/btree?status.svg)](https://godoc.org/github.com/tidwall/btree)

An [efficient](#performance) [B-tree](https://en.wikipedia.org/wiki/B-tree) implementation in Go. 

*Check out the [generics branch](https://github.com/tidwall/btree/tree/generics) if you want to try out btree with generic support for Go 1.18+*

## Features

- `Copy()` method with copy-on-write support.
- Fast bulk loading for pre-ordered data using the `Load()` method.
- All operations are thread-safe.
- [Path hinting](PATH_HINT.md) optimization for operations with nearby keys.

## Installing

To start using btree, install Go and run `go get`:

```sh
$ go get -u github.com/tidwall/btree
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/tidwall/btree"
)

type Item struct {
	Key, Val string
}

// byKeys is a comparison function that compares item keys and returns true
// when a is less than b.
func byKeys(a, b interface{}) bool {
	i1, i2 := a.(*Item), b.(*Item)
	return i1.Key < i2.Key
}

// byVals is a comparison function that compares item values and returns true
// when a is less than b.
func byVals(a, b interface{}) bool {
	i1, i2 := a.(*Item), b.(*Item)
	if i1.Val < i2.Val {
		return true
	}
	if i1.Val > i2.Val {
		return false
	}
	// Both vals are equal so we should fall though
	// and let the key comparison take over.
	return byKeys(a, b)
}

func main() {
	// Create a tree for keys and a tree for values.
	// The "keys" tree will be sorted on the Keys field.
	// The "values" tree will be sorted on the Values field.
	keys := btree.New(byKeys)
	vals := btree.New(byVals)

	// Create some items.
	users := []*Item{
		&Item{Key: "user:1", Val: "Jane"},
		&Item{Key: "user:2", Val: "Andy"},
		&Item{Key: "user:3", Val: "Steve"},
		&Item{Key: "user:4", Val: "Andrea"},
		&Item{Key: "user:5", Val: "Janet"},
		&Item{Key: "user:6", Val: "Andy"},
	}

	// Insert each user into both trees
	for _, user := range users {
		keys.Set(user)
		vals.Set(user)
	}

	// Iterate over each user in the key tree
	keys.Ascend(nil, func(item interface{}) bool {
		kvi := item.(*Item)
		fmt.Printf("%s %s\n", kvi.Key, kvi.Val)
		return true
	})

	fmt.Printf("\n")
	// Iterate over each user in the val tree
	vals.Ascend(nil, func(item interface{}) bool {
		kvi := item.(*Item)
		fmt.Printf("%s %s\n", kvi.Key, kvi.Val)
		return true
	})

	// Output:
	// user:1 Jane
	// user:2 Andy
	// user:3 Steve
	// user:4 Andrea
	// user:5 Janet
	// user:6 Andy
	//
	// user:4 Andrea
	// user:2 Andy
	// user:6 Andy
	// user:1 Jane
	// user:5 Janet
	// user:3 Steve
}
```

## Operations

### Basic

```
Get(item)               # get an existing item
Set(item)               # insert or replace an existing item
Delete(item)            # delete an item
Len()                   # return the number of items in the btree
```

### Iteration

```
Ascend(pivot, iter)     # scan items in ascending order starting at pivot.
Descend(pivot, iter)    # scan items in descending order starting at pivot.
Iter()                  # returns a read-only iterator for for-loops.
```

### Queues

```
Min()                   # return the first item in the btree
Max()                   # return the last item in the btree
PopMin()                # remove and return the first item in the btree
PopMax()                # remove and return the last item in the btree
```
### Bulk loading

```
Load(item)              # load presorted items into tree
```

### Path hints

```
SetHint(item, *hint)    # insert or replace an existing item
GetHint(item, *hint)    # get an existing item
DeleteHint(item, *hint) # delete an item
```

### Array-like operations

```
GetAt(index)     # returns the value at index
DeleteAt(index)  # deletes the item at index
```

## Performance

This implementation was designed with performance in mind. 

The following benchmarks were run on my 2019 Macbook Pro (2.4 GHz 8-Core Intel Core i9) using Go 1.17.3. The items are simple 8-byte ints. 

- `google`: The [google/btree](https://github.com/google/btree) package
- `tidwall`: The [tidwall/btree](https://github.com/tidwall/btree) package
- `go-arr`: Just a simple Go array

```
** sequential set **
google:  set-seq        1,000,000 ops in 178ms, 5,618,049/sec, 177 ns/op, 39.0 MB, 40 bytes/op
tidwall: set-seq        1,000,000 ops in 156ms, 6,389,837/sec, 156 ns/op, 23.5 MB, 24 bytes/op
tidwall: set-seq-hint   1,000,000 ops in 78ms, 12,895,355/sec, 77 ns/op, 23.5 MB, 24 bytes/op
tidwall: load-seq       1,000,000 ops in 53ms, 18,937,400/sec, 52 ns/op, 23.5 MB, 24 bytes/op
go-arr:  append         1,000,000 ops in 78ms, 12,843,432/sec, 77 ns/op

** random set **
google:  set-rand       1,000,000 ops in 555ms, 1,803,133/sec, 554 ns/op, 29.7 MB, 31 bytes/op
tidwall: set-rand       1,000,000 ops in 545ms, 1,835,818/sec, 544 ns/op, 29.6 MB, 31 bytes/op
tidwall: set-rand-hint  1,000,000 ops in 670ms, 1,493,473/sec, 669 ns/op, 29.6 MB, 31 bytes/op
tidwall: set-again      1,000,000 ops in 681ms, 1,469,038/sec, 680 ns/op
tidwall: set-after-copy 1,000,000 ops in 670ms, 1,493,230/sec, 669 ns/op
tidwall: load-rand      1,000,000 ops in 569ms, 1,756,187/sec, 569 ns/op, 29.6 MB, 31 bytes/op

** sequential get **
google:  get-seq        1,000,000 ops in 165ms, 6,048,307/sec, 165 ns/op
tidwall: get-seq        1,000,000 ops in 144ms, 6,940,120/sec, 144 ns/op
tidwall: get-seq-hint   1,000,000 ops in 78ms, 12,815,243/sec, 78 ns/op

** random get **
google:  get-rand       1,000,000 ops in 701ms, 1,427,507/sec, 700 ns/op
tidwall: get-rand       1,000,000 ops in 679ms, 1,473,531/sec, 678 ns/op
tidwall: get-rand-hint  1,000,000 ops in 824ms, 1,213,805/sec, 823 ns/op
```

*You can find the benchmark utility at [tidwall/btree-benchmark](https://github.com/tidwall/btree-benchmark)*

## Contact

Josh Baker [@tidwall](http://twitter.com/tidwall)

## License

Source code is available under the MIT [License](/LICENSE).
