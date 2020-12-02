package main

import (
	"fmt"
	"time"
)

type DomainEntry struct {
	Domain   string `json:"domain"`
	IsBounce bool   `json:"bounced"`
	C        int    `json:"c"`
}

func (d DomainEntry) String() string {
	bounce := "B"
	if !d.IsBounce {
		bounce = "D"
	}
	return fmt.Sprintf("%s:%s:%d", bounce, d.Domain, d.C)
}

type StatsQuery struct {
	Domain string `json:"domain"`
	Total  int    `json:"total"`
}

func (r StatsQuery) String() string {
	return fmt.Sprintf("%s:%d", r.Domain, r.Total)
}

type DomainStatus struct {
	Domain string `json:"domain"`
	Status string `json:"status"`
}

type QueryResponse struct {
	Domain  string `json:"domain"`
	Bounced bool   `json:"bounced"`
	Total   int    `json:"total"`
}

func (r QueryResponse) String() string {
	return fmt.Sprintf("%s:%v:%d", r.Domain, r.Bounced, r.Total)
}

type HostEntry struct {
	Host string    `json:"host"`
	Seen time.Time `json:"lastseen"`
}

func (e HostEntry) String() string {
	return fmt.Sprintf("%s:%s", e.Host, e.Seen)
}
