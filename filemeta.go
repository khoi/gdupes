package main

import (
	"fmt"
	"os"
)

type FileMeta struct {
	fileInfo os.FileInfo
	path     string
	checksum []byte
}

func (f FileMeta) String() string {
	return fmt.Sprintf("%s %x", f.path, f.checksum)
}
