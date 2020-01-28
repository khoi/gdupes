package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const tmppath = "/Users/khoi/Downloads/1"

func walk(root string, input chan<- string) {
	defer close(input)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		input <- path

		return nil
	})
}

func main() {
	fmt.Printf("Scanning %s\n", tmppath)

	input := make(chan string)

	for path := range input {
		fmt.Println(path)
	}

	fmt.Println("Done")
}
