package gocrawler

import (
	"sync"
	"net/url"
)

// Crawler struct
type Crawler struct {
	uri             string
	uriURL          *url.URL
	user            string
	password        string
	header          map[string]string
	timeout         int
	maxDepth        int64
	curDepth        int64
	maxWorkers      int
	curWorkers      int
	queue           map[string]int64
	loadedQueue     map[string]bool
	workersMux      sync.Mutex
	queueMux        sync.Mutex
	isProcessingMux sync.Mutex
	callbackChan    chan *HTTPResponse
	callbackFunc    func(*HTTPResponse)
	isProcessing    bool
}

// New Crawler
func New(uri string, user string, password string) (*Crawler) {
	c := new(Crawler)
	c.uri = uri
	c.user = user
	c.password = password
	return c
}

// SetCallbackChan set callback chan
func (c *Crawler) SetCallbackChan(callbackChan chan *HTTPResponse) {
	c.callbackChan = callbackChan
}

// SetCallbackFunc set callback function
func (c *Crawler) SetCallbackFunc(callbackFunc func(*HTTPResponse)) {
	c.callbackFunc = callbackFunc
}

// SetHeader set HTTP header
func (c *Crawler) SetHeader(header map[string]string) {
	c.header = header
}

// SetTimeout set HTTP timeout in seconds
func (c *Crawler) SetTimeout(timeout int) {
	c.timeout = timeout
}

// SetMaxDepth set max crawling depth
func (c *Crawler) SetMaxDepth(maxDepth int64) {
	c.maxDepth = maxDepth
}

// SetMaxWorkers set maximum number of threads to load
func (c *Crawler) SetMaxWorkers(maxWorkers int) {
	c.maxWorkers = maxWorkers
}

// IsProcessing return loading status
func (c *Crawler) IsProcessing() (bool) {
	c.isProcessingMux.Lock()
	defer c.isProcessingMux.Unlock()
	return c.isProcessing
}

func (c *Crawler) startProcessing() {
	c.isProcessingMux.Lock()
	defer c.isProcessingMux.Unlock()
	c.isProcessing = true
}

func (c *Crawler) stopProcessing() {
	c.isProcessingMux.Lock()
	defer c.isProcessingMux.Unlock()
	c.isProcessing = false
}

// Start start download
func (c *Crawler) Start() (error) {
	var err error
	c.uriURL, err = url.Parse(c.uri)
	if err != nil {
		return err
	}

	c.queue = make(map[string]int64, 100)
	c.loadedQueue = make(map[string]bool, 100)
	c.startProcessing()

	c.addToQueue(c.uri, 0)
	go c.loop()

	return nil
}

// Stop stop download
func (c *Crawler) Stop() {
	c.stopProcessing()
}

func (c *Crawler) loop() {
	for {
		if !c.IsProcessing() {
			break
		}
		c.queueMux.Lock()
		for uri, item := range c.queue {
			if c.getCurrentWorkers() < c.maxWorkers {
				delete(c.queue, uri)
				c.startWorker(uri, item)
			}
		}
		if len(c.queue) == 0 && c.getCurrentWorkers() == 0 {
			c.stopProcessing()
			c.queueMux.Unlock()
			return
		}
		c.queueMux.Unlock()
	}
}

