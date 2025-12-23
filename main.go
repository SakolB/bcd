// bcd is intended as an improved version of cd
// where long path can be searched, and extra
// features are added to make cd less painful
// between long relative or absolute paths
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func walkUp(path string) []string {
	var parents []string
	for {
		parent := filepath.Dir(path)
		parents = append(parents, path)
		if parent == path {
			break
		}
		path = parent
	}
	return parents
}

func walkDown(root string) ([]string, error) {
	// Example: only directories, skip the root itself.
	// If you want to include root, drop -mindepth 1.
	cmd := exec.Command("find", root, "-mindepth", "1", "-type", "d")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("walkDown: stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("walkDown: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("walkDown: start find: %w", err)
	}

	var all []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		path := scanner.Text()
		all = append(all, path)
	}

	// Drain stderr so find can exit cleanly
	_, _ = bufio.NewReader(stderr).ReadString(0)

	if scanErr := scanner.Err(); scanErr != nil {
		return all, fmt.Errorf("walkDown: scan: %w", scanErr)
	}

	if err := cmd.Wait(); err != nil {
		return all, fmt.Errorf("walkDown: find failed: %w", err)
	}

	return all, nil
}

func walkAll(path string) ([]string, error) {
	parents := walkUp(path)

	var (
		allDir   []string
		mu       sync.Mutex
		wg       sync.WaitGroup
		firstErr error
		errOnce  sync.Once
	)

	// Helper to record the first error
	setErr := func(err error) {
		if err == nil {
			return
		}
		errOnce.Do(func() {
			firstErr = err
		})
	}

	// Walk all parents concurrently
	for _, parent := range parents {
		p := parent // capture loop variable
		wg.Add(1)
		go func() {
			defer wg.Done()
			dirs, err := walkDown(p)
			if err != nil {
				setErr(err)
				return
			}
			mu.Lock()
			allDir = append(allDir, dirs...)
			mu.Unlock()
		}()
	}

	// Also walk the original path concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		dirs, err := walkDown(path)
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		allDir = append(allDir, dirs...)
		mu.Unlock()
	}()

	wg.Wait()

	if firstErr != nil {
		return allDir, firstErr
	}
	return allDir, nil
}

func main() {
	dir, _ := os.Getwd()
	paths, err := walkAll(dir)
	if err != nil {
		fmt.Println("Error")
		fmt.Println(err)
	}
	for _, path := range paths {
		fmt.Println(path)
	}
	// var relativePaths []string
	// for _, path := range paths {
	// 	relativePath, err := filepath.Rel(dir, path)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	relativePaths = append(relativePaths, relativePath)
	// }
	// for _, path := range relativePaths {
	// 	fmt.Printf("Base path: %s, Relative path: %s\n", dir, path)
	// }
}
