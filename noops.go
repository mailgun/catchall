package main

import "errors"

type NoopBackend struct {
	FailGet     bool
	FailUpdate  bool
	FailInsert  bool
	FailPing    bool
	IsDelayFlag bool
	GetResp     interface{}
}

func (b NoopBackend) get(domain string) (QueryResponse, error) {
	if b.FailGet {
		return QueryResponse{}, errors.New("simulated failure")
	}
	ret := QueryResponse{}
	if b.GetResp != nil {
		ret = b.GetResp.(QueryResponse)
	}
	return ret, nil
}

func (b NoopBackend) update() ([]HostEntry, error) {
	if b.FailUpdate {
		return nil, errors.New("simulated failure")
	}
	ret := make([]HostEntry, 0)
	return ret, nil
}

func (b NoopBackend) insert(entry DomainEntry) error {
	if b.FailInsert {
		return errors.New("simulated failure")
	}
	return nil
}

func (b NoopBackend) ping(entry HostEntry) error {
	if b.FailPing {
		return errors.New("simulated failure")
	}
	return nil
}

func (b NoopBackend) isDelayed() bool {
	return b.IsDelayFlag
}

func (b NoopBackend) delay(seconds int) {
	return
}
