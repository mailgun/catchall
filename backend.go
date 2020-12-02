package main

import (
	"errors"
	"log"
	"sync"
)

type Backend interface {
	isDelayed() bool
	delay(int)
	insert(DomainEntry) error
	ping(HostEntry) error
	get(string) (QueryResponse, error)
	update() ([]HostEntry, error)
}

// distributedBackend
type distributedBackend struct {
	hosts        []string
	mu           sync.RWMutex
	persistance  Backend
	nodeIntercom NodeIntercom
}

func NewDistributedBackend(
	hosts []string, persistBackend Backend, intercom NodeIntercom) (
	*distributedBackend, error) {
	if persistBackend == nil {
		return nil, errors.New(
			"Require non-nil backend to instantiate this object")
	}
	if len(hosts) == 0 {
		return nil, errors.New(
			"This object requires at least one host (itself)")
	}
	b := new(distributedBackend)
	b.hosts = hosts
	b.persistance = persistBackend
	b.nodeIntercom = intercom
	return b, nil
}

func (b *distributedBackend) readHosts() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var ret = make([]string, len(b.hosts))
	copy(ret, b.hosts)
	return ret
}

func (b *distributedBackend) writeHosts(hosts []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	copy(b.hosts, hosts)
}

func (b *distributedBackend) ping(entry HostEntry) error {
	return b.persistance.ping(entry)
}

func (b *distributedBackend) get(domain string) (QueryResponse, error) {
	var total int = 0
	bounced := false
	if len(domain) == 0 {
		return QueryResponse{}, nil
	}
	resp, err := b.persistance.get(domain)
	if err != nil {
		log.Printf("Error selecting from persistance: %s\n", err)
	} else if resp.Bounced {
		ret := QueryResponse{Bounced: true, Total: 0, Domain: domain}
		return ret, nil
	}
	total += resp.Total // Note: resp.Total defaults to zero
	hosts := b.readHosts()
	for _, host := range hosts {
		if len(host) == 0 { // if host is an empty string
			continue
		}
		uri := "http://" + host + "/stats/" + domain
		statsQuery := QueryResponse{}
		// could probably communicate locally in a better way
		err := b.nodeIntercom.Get(uri, &statsQuery)
		if err != nil {
			log.Printf("Failed stats query: %v\n", err)
			continue
		}
		if statsQuery.Total < 0 {
			bounced = true
			break
		} else {
			// could break out of loop if total > cutoff
			total += statsQuery.Total
		}
	}
	ret := QueryResponse{Bounced: bounced, Total: total, Domain: domain}
	return ret, nil
}

func (b *distributedBackend) update() ([]HostEntry, error) {
	results, err := b.persistance.update()
	var newHosts []string = make([]string, len(results))
	for _, host := range results {
		newHosts = append(newHosts, host.Host)
	}
	b.writeHosts(newHosts)
	return results, err
}

func (b *distributedBackend) insert(entry DomainEntry) error {
	return nil
}

func (b *distributedBackend) isDelayed() bool {
	return false
}

func (b *distributedBackend) delay(seconds int) {
	return
}
