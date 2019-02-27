package gocrawler

import (
	"net/http"
	"io"
	"time"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
)

// HTTPResponse struct
type HTTPResponse struct {
	URI           string
	StatusCode    int
	ContentLength int64
	Header        http.Header
	Body          []byte
}

func httpGET(uri string, user string, password string, header map[string]string, timeout int) (*HTTPResponse, error) {
	client := &http.Client{
		Timeout:       time.Duration(timeout) * time.Second,
		CheckRedirect: nil,
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	if user != "" && password != "" {
		req.SetBasicAuth(user, password)
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	var resp = &http.Response{}

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				return nil, err
			}
		default:
			reader = resp.Body
		}
		defer reader.Close()

		body, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		r := &HTTPResponse{
			URI:           uri,
			StatusCode:    resp.StatusCode,
			ContentLength: resp.ContentLength,
			Header:        resp.Header,
			Body:          body,
		}

		return r, nil
	}
	return &HTTPResponse{StatusCode: resp.StatusCode}, fmt.Errorf("HTTP error: %d", resp.StatusCode)
}

func prepareURI(curURL, newURL *url.URL) (string) {
	if newURL.Host != "" && newURL.Host != curURL.Host {
		return ""
	}

	scheme := newURL.Scheme
	host := newURL.Host
	if scheme == "" {
		scheme = curURL.Scheme
	}
	if host == "" {
		host = curURL.Host
	}

	if !strings.HasPrefix(scheme, "http") {
		return ""
	}

	u := &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     newURL.EscapedPath(),
		RawQuery: newURL.RawQuery,
	}

	return u.String()
}
