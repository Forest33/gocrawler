package gocrawler

import (
	"sync"
	"net/url"
)

// Crawler struct
type Crawler struct {
	uri                  string
	uriURL               *url.URL
	user                 string
	password             string
	header               map[string]string
	timeout              int
	maxDepth             int64
	curDepth             int64
	maxWorkers           int
	curWorkers           int
	isLoadImages         bool
	queue                map[string]int64
	loadedQueue          map[string]bool
	loadedImages         map[string]bool
	imageWorkersChan     []chan *imageRequest
	imageWorkersStopChan []chan bool
	imageWorkers         int
	workersMux           sync.Mutex
	queueMux             sync.Mutex
	isProcessingMux      sync.Mutex
	imagesMux            sync.Mutex
	callbackChan         chan *CrawlerResponse
	callbackFunc         func(*CrawlerResponse)
	isProcessing         bool
}

// Image struct
type Image struct {
	URI     string
	Payload *HTTPResponse
	Err     error
}

// imageRequest struct
type imageRequest struct {
	url  string
	cbCh chan *Image
}

// CrawlerResponse struct
type CrawlerResponse struct {
	URI     string
	Payload *HTTPResponse
	Images  []*Image
	Err     error
}

// New Crawler
func New(uri string, user string, password string) (*Crawler) {
	c := new(Crawler)
	c.uri = uri
	c.user = user
	c.password = password
	c.maxWorkers = 1
	c.imageWorkers = 1
	return c
}

// SetCallbackChan set callback chan
func (c *Crawler) SetCallbackChan(callbackChan chan *CrawlerResponse) {
	c.callbackChan = callbackChan
}

// SetCallbackFunc set callback function
func (c *Crawler) SetCallbackFunc(callbackFunc func(*CrawlerResponse)) {
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

// SetLoadImages load or not images
func (c *Crawler) SetLoadImages(isLoadImages bool) {
	c.isLoadImages = isLoadImages
}

// SetImagesWorkers set number of threads to load images
func (c *Crawler) SetImagesWorkers(imageWorkers int) {
	c.imageWorkers = imageWorkers
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

	if c.isLoadImages {
		for i := 0; i < c.imageWorkers; i ++ {
			c.imageWorkersStopChan[i] <- true
		}
	}

	c.isProcessing = false
}

// Start start download
func (c *Crawler) Start() (error) {
	var err error
	c.uriURL, err = url.Parse(c.uri)
	if err != nil {
		return err
	}

	if c.isLoadImages {
		for i := 0; i < c.imageWorkers; i ++ {
			ch := make(chan *imageRequest, c.imageWorkers*10)
			stopCh := make(chan bool)
			c.imageWorkersChan = append(c.imageWorkersChan, ch)
			c.imageWorkersStopChan = append(c.imageWorkersStopChan, stopCh)
			go c.imageWorker(ch, stopCh)
		}
	}

	c.queue = make(map[string]int64, 100)
	c.loadedQueue = make(map[string]bool, 1000)
	c.loadedImages = make(map[string]bool, 1000)
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
	c.stopProcessing()
}
