package conv

import (
	"unsafe"
)

func StringToBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	// https://github.com/golang/go/blob/go1.27.1/src/os/file.go#L320
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// https://github.com/golang/go/blob/go1.27.1/src/strings/builder.go#L47
	return unsafe.String(unsafe.SliceData(b), len(b))
}
