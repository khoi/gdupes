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

func (fileMeta FileMeta) String() string {
	return fmt.Sprintf("%s %x", fileMeta.path, fileMeta.checksum)
}
