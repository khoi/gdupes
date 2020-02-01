package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

func groupFilesBySize(ctx context.Context, roots ...string) map[int64][]*FileMeta {
	result := make(map[int64][]*FileMeta)

	for v := range walkFilesInDirectories(ctx, roots...) {
		if v.err != nil {
			fmt.Fprintln(os.Stderr, v.err)
			continue
		}
		result[v.info.Size()] = append(result[v.info.Size()], &FileMeta{v.info, v.path, nil})
	}
	return result
}

func hashWorker(wg *sync.WaitGroup, in <-chan *FileMeta) {
	defer wg.Done()
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
	}
}

func printUsage() {
	fmt.Fprintln(os.Stdout, "gdupes - duplicate files finder")
	fmt.Fprintln(os.Stdout, "Usage: ")
	fmt.Fprintln(os.Stdout, "\tgdupes /path1 /path2")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	filesGroupedBySize := groupFilesBySize(context.TODO(), os.Args[1:]...)
	var filesNeedHashing []*FileMeta

	for _, v := range filesGroupedBySize {
		if len(v) < 2 { // skip unique group
			continue
		}
		filesNeedHashing = append(filesNeedHashing, v...)
	}

	numWorker := runtime.NumCPU()
	numJobs := len(filesNeedHashing)
	jobs := make(chan *FileMeta, numJobs)
	finishedHashing := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(numWorker)

	for i := 0; i < numWorker; i++ {
		go hashWorker(&wg, jobs)
	}

	for _, fileMeta := range filesNeedHashing {
		jobs <- fileMeta
	}
	close(jobs)

	go func() {
		wg.Wait()
		finishedHashing <- struct{}{}
	}()

	<-finishedHashing

	filesGroupedByHash := make(map[string][]*FileMeta)
	for _, f := range filesNeedHashing {
		checksum := string(f.checksum)
		filesGroupedByHash[checksum] = append(filesGroupedByHash[checksum], f)
	}

	// output duplicated
	for _, files := range filesGroupedByHash {
		for _, f := range files {
			fmt.Fprintf(os.Stdout, "%s\n", f.path)
		}
		fmt.Fprintln(os.Stdout)
	}
}
