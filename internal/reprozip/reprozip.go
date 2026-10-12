package reprozip

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
)

type Entry struct {
	Name string
	Data []byte
	Mode uint32
}

type encodedEntry struct {
	Entry
	Compressed []byte
	CRC        uint32
	Offset     uint32
}

func put16(b *bytes.Buffer, v uint16) { _ = binary.Write(b, binary.LittleEndian, v) }
func put32(b *bytes.Buffer, v uint32) { _ = binary.Write(b, binary.LittleEndian, v) }

// WritePythonCompatible writes exactly the ZIP layout produced by the historical
// Python zipfile helper used by this project: Unix creator, deflate, 2020-01-01
// timestamp, 0644 regular-file mode, no data descriptors, no extras/comments.
// The compressor is zlib's raw DEFLATE with Python's defaults, implemented in
// deflate_cgo.go so the compressed streams are byte-identical too.
func WritePythonCompatible(path string, entries []Entry) error {
	if len(entries) > 0xffff {
		return fmt.Errorf("too many ZIP entries: %d", len(entries))
	}
	enc := make([]encodedEntry, len(entries))
	var out bytes.Buffer

	for i, e := range entries {
		if e.Mode == 0 {
			e.Mode = 0o644
		}
		if len(e.Name) > 0xffff {
			return fmt.Errorf("ZIP member name too long: %q", e.Name)
		}
		comp, err := rawDeflate(e.Data)
		if err != nil {
			return fmt.Errorf("deflate %s: %w", e.Name, err)
		}
		ee := encodedEntry{Entry: e, Compressed: comp, CRC: crc32.ChecksumIEEE(e.Data), Offset: uint32(out.Len())}
		enc[i] = ee

		put32(&out, 0x04034b50)
		put16(&out, 20)     // version needed
		put16(&out, 0)      // flags
		put16(&out, 8)      // deflate
		put16(&out, 0)      // DOS time 00:00:00
		put16(&out, 0x5021) // DOS date 2020-01-01
		put32(&out, ee.CRC)
		put32(&out, uint32(len(comp)))
		put32(&out, uint32(len(e.Data)))
		put16(&out, uint16(len(e.Name)))
		put16(&out, 0)
		out.WriteString(e.Name)
		out.Write(comp)
	}

	centralOffset := out.Len()
	for _, e := range enc {
		put32(&out, 0x02014b50)
		put16(&out, 0x0314) // Unix creator, ZIP 2.0
		put16(&out, 20)
		put16(&out, 0)
		put16(&out, 8)
		put16(&out, 0)
		put16(&out, 0x5021)
		put32(&out, e.CRC)
		put32(&out, uint32(len(e.Compressed)))
		put32(&out, uint32(len(e.Data)))
		put16(&out, uint16(len(e.Name)))
		put16(&out, 0)                   // extra
		put16(&out, 0)                   // comment
		put16(&out, 0)                   // disk
		put16(&out, 0)                   // internal attrs
		put32(&out, (0x8000|e.Mode)<<16) // S_IFREG | mode
		put32(&out, e.Offset)
		out.WriteString(e.Name)
	}
	centralSize := out.Len() - centralOffset

	put32(&out, 0x06054b50)
	put16(&out, 0)
	put16(&out, 0)
	put16(&out, uint16(len(enc)))
	put16(&out, uint16(len(enc)))
	put32(&out, uint32(centralSize))
	put32(&out, uint32(centralOffset))
	put16(&out, 0)

	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out.Bytes(), 0o644)
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			if i == 0 {
				return string(path[0])
			}
			return path[:i]
		}
	}
	return "."
}
