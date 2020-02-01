package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
)

type walkResult struct {
	path string
	info os.FileInfo
	err  error
}

func walkFiles(ctx context.Context, root string) <-chan *walkResult {
	result := make(chan *walkResult)
	go func() {
		defer close(result)
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				result <- &walkResult{
					path: path,
					info: nil,
					err:  err,
				}
				return nil
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			select {
			case result <- &walkResult{path, info, nil}:
			case <-ctx.Done():
				return nil
			}

			return nil
		})
	}()

	return result
}

func walkFilesInDirectories(ctx context.Context, roots ...string) <-chan *walkResult {
	result := make(chan *walkResult, len(roots)*100)

	var wg sync.WaitGroup
	wg.Add(len(roots))

	f := func(res <-chan *walkResult) {
		defer wg.Done()
		for r := range res {
			result <- r
		}
	}

	for _, root := range roots {
		go f(walkFiles(ctx, root))
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}
