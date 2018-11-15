package crawler

import (
	"sync"
	"net/url"
)

type Crawler struct {
	uri             string
	uriUrl          *url.URL
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
	callbackChan    chan *HttpResponse
	callbackFunc    func(*HttpResponse)
	isProcessing    bool
}

func New(uri string, user string, password string) (*Crawler) {
	c := new(Crawler)
	c.uri = uri
	c.user = user
	c.password = password
	return c
}

func (self *Crawler) SetCallbackChan(callbackChan chan *HttpResponse) {
	self.callbackChan = callbackChan
}

func (self *Crawler) SetCallbackFunc(callbackFunc func(*HttpResponse)) {
	self.callbackFunc = callbackFunc
}

func (self *Crawler) SetHeader(header map[string]string) {
	self.header = header
}

func (self *Crawler) SetTimeout(timeout int) {
	self.timeout = timeout
}

func (self *Crawler) SetMaxDepth(maxDepth int64) {
	self.maxDepth = maxDepth
}

func (self *Crawler) SetMaxWorkers(maxWorkers int) {
	self.maxWorkers = maxWorkers
}

func (self *Crawler) IsProcessing() (bool) {
	self.isProcessingMux.Lock()
	defer self.isProcessingMux.Unlock()
	return self.isProcessing
}

func (self *Crawler) startProcessing() {
	self.isProcessingMux.Lock()
	defer self.isProcessingMux.Unlock()
	self.isProcessing = true
}

func (self *Crawler) stopProcessing() {
	self.isProcessingMux.Lock()
	defer self.isProcessingMux.Unlock()
	self.isProcessing = false
}

func (self *Crawler) Start() (error) {
	var err error
	self.uriUrl, err = url.Parse(self.uri)
	if err != nil {
		return err
	}

	self.queue = make(map[string]int64, 100)
	self.loadedQueue = make(map[string]bool, 100)
	self.startProcessing()

	self.addToQueue(self.uri, 0)
	go self.loop()

	return nil
}

func (self *Crawler) loop() {
	for {
		self.queueMux.Lock()
		for uri, item := range self.queue {
			if self.getCurrentWorkers() < self.maxWorkers {
				delete(self.queue, uri)
				self.startWorker(uri, item)
			}
		}
		if len(self.queue) == 0 && self.getCurrentWorkers() == 0 {
			self.stopProcessing()
			self.queueMux.Unlock()
			return
		}
		self.queueMux.Unlock()
	}
}

