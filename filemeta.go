package main

import (
	"fmt"
)

type FileMeta struct {
	path     string
	checksum []byte
}

func (f FileMeta) String() string {
	return fmt.Sprintf("%s %x", f.path, f.checksum)
}
