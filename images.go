package gocrawler

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
	"net/url"
)

func (c *Crawler) loadImages(doc *goquery.Document) ([]*Image, error) {
	var imagesURL []string

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		uri, exists := s.Attr("src")
		if exists {
			uri = strings.TrimSpace(uri)
			u, err := url.Parse(uri)
			if err == nil {
				link := prepareURI(c.uriURL, u, false)
				if link != "" {
					imagesURL = append(imagesURL, link)
				}
			}
		}
	})

	if len(imagesURL) > 0 {
		images := make([]*Image, 0, len(imagesURL))
		cbCh := make(chan *Image, len(imagesURL))

		workerIdx := 0
		for _, img := range imagesURL {
			if workerIdx >= c.imageWorkers {
				workerIdx = 0
			}
			c.imageWorkersChan[workerIdx] <- &imageRequest{
				url:  img,
				cbCh: cbCh,
			}
			workerIdx ++
		}

		loadedCount := 0
		for {
			img := <-cbCh
			images = append(images, img)
			loadedCount ++
			if loadedCount >= len(imagesURL) {
				break
			}
		}

		return images, nil
	}

	return []*Image{}, nil
}

func (c *Crawler) imageWorker(ch chan *imageRequest, stopCh chan bool) {
	isProcessing := true
	for isProcessing {
		select {
		case request := <-ch:
			c.imagesMux.Lock()
			if _, ok := c.loadedImages[request.url]; !ok {
				response, err := httpGET(request.url, c.user, c.password, c.header, c.timeout)
				if err == nil {
					c.loadedImages[request.url] = true
				}
				request.cbCh <- &Image{
					URI:     request.url,
					Payload: response,
					Err:     err,
				}
			} else {
				request.cbCh <- &Image{URI: request.url}
			}
			c.imagesMux.Unlock()
		case <-stopCh:
			isProcessing = false
			break
		}
	}
}
