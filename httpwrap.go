package main

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type HTTPGetter interface {
	Get(string) (*http.Response, error)
}

type DefaultHTTP struct{}

func (h *DefaultHTTP) Get(uri string) (*http.Response, error) {
	return http.Get(uri)
}

type MockedHTTP struct {
	Ret        string
	StatusCode int
	Err        error
	Called     bool
}

func (h *MockedHTTP) Get(uri string) (*http.Response, error) {
	h.Called = true
	resp := http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(h.Ret)),
		StatusCode: h.StatusCode,
	}
	return &resp, h.Err
}
