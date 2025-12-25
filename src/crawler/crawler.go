// Package crawler implements directory crawler that
// traverses all directory from a starting directory
// using BFS as supposed to the the filepath DFS traversal
package crawler

import (
	"os"
	"path/filepath"
)

type Crawler struct {
	pathChan     chan string
	errChan      chan error
	done         chan struct{}
	skipHidden   bool
	maxDepth     int
	ignoreErrors bool
}

func NewCrawler() *Crawler {
	return &Crawler{
		pathChan:     make(chan string, 1000),
		errChan:      make(chan error, 20),
		done:         make(chan struct{}),
		skipHidden:   false,
		maxDepth:     -1,
		ignoreErrors: true,
	}
}

func (c *Crawler) Paths() <-chan string {
	return c.pathChan
}

func (c *Crawler) Errors() <-chan error {
	return c.errChan
}

func (c *Crawler) Done() <-chan struct{} {
	return c.done
}

// Crawl crawls the directory from basedDir
// using BFS traversal. Any error, encountered are
// sent on the c.errChan channel. All path
// discovered will be send to c.pathChan channel.
func (c *Crawler) Crawl(baseDir string) {
	defer close(c.pathChan)
	defer close(c.errChan)
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		if !c.ignoreErrors {
			c.errChan <- err
		}
		return
	}
	queue := make([]string, 0)
	visited := make(map[string]bool)
	queue = append(queue, absDir)
	visited[absDir] = true
	c.pathChan <- absDir
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		neighbors, err := c.getNeighbor(current)
		if err != nil {
			if !c.ignoreErrors {
				c.errChan <- err
			}
		}
		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
				c.pathChan <- neighbor
			}
		}
	}
}

// getNeigbor take a directory path (absolute path)
// returns a list of string containing its neighbor directories, and
// an error. Neighbor directories are any directory that is either a direct
// child and parent directory. It passes the path of any children
// entries that are files into crawler's pathChan channel.
func (c *Crawler) getNeighbor(dir string) ([]string, error) {
	var neighbors []string
	children, err := os.ReadDir(dir)
	for _, child := range children {
		childPath := filepath.Join(dir, child.Name())
		if child.IsDir() {
			neighbors = append(neighbors, childPath)
		} else {
			c.pathChan <- childPath
		}
	}
	parent := filepath.Dir(dir)
	neighbors = append(neighbors, parent)
	return neighbors, err
}
