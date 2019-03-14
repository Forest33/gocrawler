package main

import (
	"github.com/Forest33/gocrawler"
	"fmt"
	"time"
)

func init() {

}

func main() {
	uri := "https://gobyexample.com/"
	c := gocrawler.New(uri, "", "")
	ch := make(chan *gocrawler.CrawlerResponse)
	c.SetCallbackChan(ch)
	c.SetCallbackFunc(cb)
	c.SetTimeout(10)
	c.SetMaxDepth(0)
	c.SetMaxWorkers(5)
	c.SetLoadImages(true)
	c.SetImagesWorkers(5)
	c.Start()

	for {
		select {
		case r := <-ch:
			fmt.Println("chan:", r.URI, r.Payload.Header.Get("Content-Type"), r.Err, len(r.Images))
		case <-time.After(time.Second):
		}
		if !c.IsProcessing() {
			break
		}
	}
}

func cb(response *gocrawler.CrawlerResponse) {
	fmt.Println("func:", response.URI, response.Payload.Header.Get("Content-Type"), response.Err, len(response.Images))
}
