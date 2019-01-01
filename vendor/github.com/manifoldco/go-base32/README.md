# go-base32

base32 encoding using Manifold's chosen alphabet, with no padding.

[Code of Conduct](./.github/CONDUCT.md) |
[Contribution Guidelines](./.github/CONTRIBUTING.md)

[![GitHub release](https://img.shields.io/github/tag/manifoldco/go-base32.svg?label=latest)](https://github.com/manifoldco/go-base32/releases)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/manifoldco/go-base32)
[![Travis](https://img.shields.io/travis/manifoldco/go-base32/master.svg)](https://travis-ci.org/manifoldco/go-base32)
[![Go Report Card](https://goreportcard.com/badge/github.com/manifoldco/go-base32)](https://goreportcard.com/report/github.com/manifoldco/go-base32)
[![License](https://img.shields.io/badge/license-BSD-blue.svg)](./LICENSE.md)

## Usage

```go
package main

import (
	"fmt"

	"github.com/manifoldco/go-base32"
)

func main() {
	sample := []byte{0xF, 0xEE, 0xD, 0xC, 0x0D}

	fmt.Println(base32.EncodeToString(sample))
}
```

Outputs:

```
1zq0u30d
```
