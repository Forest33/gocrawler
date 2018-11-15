package crawler

import (
	"sync/atomic"
	"github.com/PuerkitoBio/goquery"
	"bytes"
	"strings"
	"net/url"
)

func (self *Crawler) addToQueue(uri string, depth int64) {
	self.queueMux.Lock()
	defer self.queueMux.Unlock()

	if _, ok := self.queue[uri]; ok {
		return
	}

	if self.maxDepth > 0 && depth > self.maxDepth {
		return
	}

	if _, ok := self.loadedQueue[uri]; !ok {
		self.queue[uri] = depth
		atomic.StoreInt64(&self.curDepth, depth)
	}
}

func (self *Crawler) startWorker(uri string, depth int64) {
	self.workersMux.Lock()
	self.curWorkers ++
	self.workersMux.Unlock()

	go self.worker(uri, depth)
}

func (self *Crawler) stopWorker() {
	self.workersMux.Lock()
	defer self.workersMux.Unlock()
	self.curWorkers --
}

func (self *Crawler) getCurrentWorkers() (int) {
	self.workersMux.Lock()
	defer self.workersMux.Unlock()
	return self.curWorkers
}

func (self *Crawler) worker(uri string, depth int64) {
	defer func() {
		self.stopWorker()

		self.queueMux.Lock()
		defer self.queueMux.Unlock()
		self.loadedQueue[uri] = true
	}()

	response, err := httpGET(uri, self.user, self.password, self.header, self.timeout)
	if err != nil {
		return
	}

	if self.callbackChan != nil {
		self.callbackChan <- response
	}

	if self.callbackFunc != nil {
		self.callbackFunc(response)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(response.Body))
	if err != nil {
		return
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		uri, exists := s.Attr("href")
		if exists {
			uri = strings.TrimSpace(uri)
			u, err := url.Parse(uri)
			if err == nil {
				link := prepareURI(self.uriUrl, u)
				if link != "" {
					self.addToQueue(link, depth + 1)
				}
			}
		}
	})
}

