//go:build cgo

package reprozip

/*
#cgo LDFLAGS: -lz
#include <stdlib.h>
#include <string.h>
#include <zlib.h>

static int evenicle_raw_deflate(const unsigned char* in, size_t inlen,
                                unsigned char** out, size_t* outlen) {
    z_stream s;
    memset(&s, 0, sizeof(s));
    int r = deflateInit2(&s, Z_DEFAULT_COMPRESSION, Z_DEFLATED,
                         -MAX_WBITS, 8, Z_DEFAULT_STRATEGY);
    if (r != Z_OK) return r;

    uLong bound = deflateBound(&s, (uLong)inlen);
    unsigned char *buf = (unsigned char*)malloc(bound ? bound : 16);
    if (!buf) {
        deflateEnd(&s);
        return Z_MEM_ERROR;
    }

    s.next_in = (Bytef*)in;
    s.avail_in = (uInt)inlen;
    s.next_out = buf;
    s.avail_out = (uInt)(bound ? bound : 16);
    r = deflate(&s, Z_FINISH);
    if (r != Z_STREAM_END) {
        free(buf);
        deflateEnd(&s);
        return r;
    }

    *out = buf;
    *outlen = (size_t)s.total_out;
    deflateEnd(&s);
    return Z_OK;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func rawDeflate(in []byte) ([]byte, error) {
	var src *C.uchar
	if len(in) != 0 {
		src = (*C.uchar)(unsafe.Pointer(&in[0]))
	}
	var out *C.uchar
	var outLen C.size_t
	rc := C.evenicle_raw_deflate(src, C.size_t(len(in)), &out, &outLen)
	if rc != C.Z_OK {
		return nil, fmt.Errorf("zlib deflate failed: %d", int(rc))
	}
	defer C.free(unsafe.Pointer(out))
	if outLen > C.size_t(^uint(0)>>1) {
		return nil, fmt.Errorf("compressed stream too large")
	}
	return C.GoBytes(unsafe.Pointer(out), C.int(outLen)), nil
}
