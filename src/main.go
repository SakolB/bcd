package main

import (
	"fmt"

	"github.com/sakolb/bcd/src/crawler"
)

func main() {
	c := crawler.NewCrawler()
	pathChan := c.Paths()
	go c.Crawl("./")
	go func() {
		for err := range c.Errors() {
			fmt.Println(err)
		}
	}()
	for path := range pathChan {
		fmt.Println(path)
	}
}
