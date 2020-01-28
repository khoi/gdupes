package main

import (
	"os"
)

type FileMeta struct {
	os.FileInfo
	path string
}

func (f FileMeta) String() string {
	return f.path
}
