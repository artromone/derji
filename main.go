package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func main() {
	var root string
	flag.StringVar(&root, "path", ".", "Путь к директории")
	flag.Parse()

	fileCh := make(chan string, 100)
	resultCh := make(chan string, 100)

	var wg sync.WaitGroup

	numWorkers := runtime.NumCPU()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for path := range fileCh {
				content, err := readFileContent(path)
				if err == nil {
					resultCh <- fmt.Sprintf("\n=== %s ===\n%s", path, content)
				}
			}
		}()
	}

	go func() {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}

			if !info.IsDir() && info.Size() <= 1*1024*1024 && !isExecutable(info) {
				fileCh <- path
			}
			return nil
		})
		close(fileCh)
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for res := range resultCh {
		fmt.Print(res)
	}
}

func readFileContent(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	reader := bufio.NewReader(file)
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func isExecutable(info os.FileInfo) bool {
	mode := info.Mode()
	return mode&0o111 != 0
}
