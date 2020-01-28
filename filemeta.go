package main

import (
	"fmt"
	"os"
)

type FileMeta struct {
	fileInfo os.FileInfo
	path     string
}

type FileWithChecksumMeta struct {
	fileMeta FileMeta
	checksum []byte
}

func (f FileWithChecksumMeta) String() string {
	return fmt.Sprintf("%s %x", f.fileMeta.path, f.checksum)
}
