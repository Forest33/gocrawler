package gocrawler

import (
	"sync/atomic"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"net/url"
	"bytes"
)

func (c *Crawler) addToQueue(uri string, depth int64) {
	c.queueMux.Lock()
	defer c.queueMux.Unlock()

	if _, ok := c.queue[uri]; ok {
		return
	}

	if c.maxDepth > 0 && depth > c.maxDepth {
		return
	}

	if _, ok := c.loadedQueue[uri]; !ok {
		c.queue[uri] = depth
		atomic.StoreInt64(&c.curDepth, depth)
	}
}

func (c *Crawler) startWorker(uri string, depth int64) {
	c.workersMux.Lock()
	c.curWorkers ++
	c.workersMux.Unlock()

	go c.worker(uri, depth)
}

func (c *Crawler) stopWorker() {
	c.workersMux.Lock()
	defer c.workersMux.Unlock()
	c.curWorkers --
}

func (c *Crawler) getCurrentWorkers() (int) {
	c.workersMux.Lock()
	defer c.workersMux.Unlock()
	return c.curWorkers
}

func (c *Crawler) worker(uri string, depth int64) {
	defer func() {
		c.stopWorker()

		c.queueMux.Lock()
		defer c.queueMux.Unlock()
		c.loadedQueue[uri] = true
	}()

	var (
		response *HTTPResponse
		doc      *goquery.Document
		err      error
	)

	response, err = httpGET(uri, c.user, c.password, c.header, c.timeout)

	if err == nil {
		doc, err = goquery.NewDocumentFromReader(bytes.NewReader(response.Body))
		if err != nil {
			return
		}
	}

	images := []*Image{}
	if err == nil && c.isLoadImages {
		images, err = c.loadImages(doc)
	}

	if c.callbackChan != nil {
		c.callbackChan <- &CrawlerResponse{
			URI:     uri,
			Payload: response,
			Images:  images,
			Err:     err,
		}
	}

	if c.callbackFunc != nil {
		c.callbackFunc(&CrawlerResponse{
			URI:     uri,
			Payload: response,
			Images:  images,
			Err:     err,
		})
	}

	if err != nil {
		return
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		uri, exists := s.Attr("href")
		if exists {
			uri = strings.TrimSpace(uri)
			u, err := url.Parse(uri)
			if err == nil {
				link := prepareURI(c.uriURL, u, true)
				if link != "" {
					c.addToQueue(link, depth+1)
				}
			}
		}
	})
}
