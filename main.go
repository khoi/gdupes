package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func walkFiles(done <-chan struct{}, root string) (<-chan *FileMeta, <-chan error) {
	res := make(chan *FileMeta)
	errorChan := make(chan error, 1)

	go func() {
		defer close(res)
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

			select {
			case res <- &FileMeta{fileInfo: info, path: path}:
			case <-done:
				return errors.New("cancelled")
			}

			return nil
		})

		errorChan <- err
	}()

	return res, errorChan
}

func digest(done <-chan struct{}, in <-chan *FileMeta, out chan<- *FileMeta) {
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

		select {
		case out <- meta:
		case <-done:
			return
		}
	}
}

func MD5All(root string, workerNum int) ([]*FileMeta, error) {
	done := make(chan struct{})
	defer close(done)

	paths, errc := walkFiles(done, root)

	c := make(chan *FileMeta)
	var wg sync.WaitGroup
	wg.Add(workerNum)

	for i := 0; i < workerNum; i++ {
		go func() {
			digest(done, paths, c)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	var res []*FileMeta

	for r := range c {
		res = append(res, r)
	}

	if err := <-errc; err != nil {
		return nil, err
	}

	return res, nil
}

func main() {
	m, err := MD5All(os.Args[1], runtime.NumCPU())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(m)
}
