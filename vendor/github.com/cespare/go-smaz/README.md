# go-smaz

[![GoDoc](https://godoc.org/github.com/cespare/go-smaz?status.svg)](https://godoc.org/github.com/cespare/go-smaz)

go-smaz is a pure Go implementation of [antirez's](https://github.com/antirez)
[smaz](https://github.com/antirez/smaz), a library for compressing short strings (particularly containing
English words).

## Installation

    $ go get github.com/cespare/go-smaz

## Usage

``` go
import (
  "github.com/cespare/go-smaz"
)

func main() {
  s := "Now is the time for all good men to come to the aid of the party."
  compressed := smaz.Compress([]byte(s))           // type is []byte
  decompressed, err := smaz.Decompress(compressed) // type is []byte; string(decompressed) == s
  if err != nil {
    ...
}
```

Also see the [API documentation](http://godoc.org/github.com/cespare/go-smaz).

## Notes

go-smaz is not a direct port of the C version. It is not guaranteed that the output of `smaz.Compress` will be
precisely the same as the C library. However, the output should be decompressible by the C library, and the
output of the C library should be decompressible by `smaz.Decompress`.

## Author

Caleb Spare ([cespare](https://github.com/cespare)). smaz was created by Salvatore Sanfilippo
([antirez](https://github.com/antirez)).

## Contributors

* [Antoine Grondin](https://github.com/aybabtme)

## License

MIT Licensed.

## Other implementations

* [The original C implementation](https://github.com/antirez/smaz)
* [Javascript](https://npmjs.org/package/smaz)
