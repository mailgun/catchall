package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type NodeIntercom interface {
	Get(string, interface{}) error
}

type hTTPIntercom struct {
	httpClient HTTPGetter
}

func NewHTTPIntercom(httpClient HTTPGetter) *hTTPIntercom {
	ic := new(hTTPIntercom)
	ic.httpClient = httpClient
	return ic
}

func (ic hTTPIntercom) Get(uri string, result interface{}) error {
	resp, err := ic.httpClient.Get(uri)
	if err != nil {
		return errors.New(fmt.Sprintf("http.Get failed: %s", err))
	}
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, result)
	if err != nil {
		return errors.New(
			fmt.Sprintf("error unmarshalling response: %s", err))
	}
	return nil
}
