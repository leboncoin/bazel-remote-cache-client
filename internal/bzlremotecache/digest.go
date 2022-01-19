package bzlremotecache

import (
	"fmt"
	"strconv"
	"strings"
)

// Digest contains the hash and the size of a blob.
type Digest struct {
	Hash string
	Size int64
}

// ParseDigestFromString parses the given digest string
// following the format <hash>/<size>.
func ParseDigestFromString(s string) (*Digest, error) {
	pair := strings.Split(s, "/")
	if len(pair) != 2 {
		return nil, fmt.Errorf("expected digest in the form hash/size, got %s", s)
	}

	size, err := strconv.ParseInt(pair[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid size in digest %s: %s", s, err)
	}

	return &Digest{
		Hash: pair[0],
		Size: size,
	}, nil
}
