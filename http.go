package crawler

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

type HttpResponse struct {
	Uri           string
	StatusCode    int
	ContentLength int64
	Header        http.Header
	Body          []byte
}

func httpGET(uri string, user string, password string, header map[string]string, timeout int) (*HttpResponse, error) {
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

		r := &HttpResponse{
			Uri:           uri,
			StatusCode:    resp.StatusCode,
			ContentLength: resp.ContentLength,
			Header:        resp.Header,
			Body:          body,
		}

		return r, nil
	} else {
		return &HttpResponse{StatusCode: resp.StatusCode}, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}
}

func prepareURI(cur_url, new_url *url.URL) (string) {
	if new_url.Host != "" && new_url.Host != cur_url.Host {
		return ""
	}

	scheme := new_url.Scheme
	host := new_url.Host
	if scheme == "" {
		scheme = cur_url.Scheme
	}
	if host == "" {
		host = cur_url.Host
	}

	if !strings.HasPrefix(scheme, "http") {
		return ""
	}

	u := &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     new_url.EscapedPath(),
		RawQuery: new_url.RawQuery,
	}

	return u.String()
}

