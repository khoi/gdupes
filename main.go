package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func walkDuplicateFiles(root string) (<-chan *FileMeta, <-chan error) {
	res := make(chan *FileMeta)
	errorChan := make(chan error, 1)

	go func() {
		defer close(res)

		groupedBySize := make(map[int64][]*FileMeta)

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "cannot read path %s", path)
				return nil
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			fileSize := info.Size()
			currentFileMeta := &FileMeta{fileInfo: info, path: path}

			groupedBySize[fileSize] = append(groupedBySize[fileSize], currentFileMeta)

			groupCount := len(groupedBySize[fileSize])

			if groupCount < 2 {
				return nil
			}

			if groupCount == 2 { // need to send first item for hash check
				res <- groupedBySize[fileSize][0]
			}

			res <- &FileMeta{fileInfo: info, path: path}

			return nil
		})

		errorChan <- err
	}()

	return res, errorChan
}

func computeHash(in <-chan *FileMeta, out chan<- *FileMeta) {
	h := md5.New()
	for meta := range in {
		f, err := os.Open(meta.path)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "cannot read file %s", meta.path)
			continue
		}

		if _, err := io.Copy(h, f); err != nil {
			fmt.Fprintln(os.Stderr, err)
			_ = f.Close()
			continue
		}

		meta.checksum = h.Sum(nil)
		_ = f.Close()
		h.Reset()

		out <- meta
	}
}

func findDuplicate(root string, workerNum int) (map[string][]*FileMeta, error) {
	paths, errc := walkDuplicateFiles(root)

	c := make(chan *FileMeta)
	var wg sync.WaitGroup
	wg.Add(workerNum)

	for i := 0; i < workerNum; i++ {
		go func() {
			computeHash(paths, c)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	result := make(map[string][]*FileMeta)

	for r := range c {
		checksum := string(r.checksum)
		result[checksum] = append(result[checksum], r)
	}

	if err := <-errc; err != nil {
		return nil, err
	}

	for k, v := range result {
		if len(v) < 2 {
			delete(result, k)
		}
	}

	return result, nil
}

func main() {
	m, err := findDuplicate(os.Args[1], runtime.NumCPU())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, v := range m {
		for _, meta := range v {
			fmt.Fprintln(os.Stdout, meta.path)
		}
		fmt.Fprintln(os.Stdout)
	}
}
