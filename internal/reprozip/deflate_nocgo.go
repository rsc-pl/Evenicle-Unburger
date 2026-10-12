//go:build !cgo

package reprozip

import "fmt"

func rawDeflate([]byte) ([]byte, error) {
	return nil, fmt.Errorf("byte-identical release ZIP creation requires a CGO-enabled builder linked with zlib")
}
