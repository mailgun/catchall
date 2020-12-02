package main

import (
	"log"
	"sync"
	"time"

	"github.com/gocql/gocql"
)

type cassandraBackend struct {
	Cluster  *gocql.ClusterConfig
	Session  *gocql.Session
	DelayEnd time.Time
	mu       sync.RWMutex
}

func (b *cassandraBackend) createSession() {
	log.Println("Starting cassandra session")
	session, err := b.Cluster.CreateSession()
	if err != nil {
		log.Fatal("Cassandra init failed")
	}
	b.Session = session
}

func (b *cassandraBackend) EndSession() {
	log.Println("Ending cassandra session")
	if b.Session != nil {
		b.Session.Close()
		b.Session = nil
	}
}

func (b *cassandraBackend) readDelay() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.DelayEnd
}

func (b *cassandraBackend) setDelay(delay time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.DelayEnd = delay
}

func (b *cassandraBackend) isDelayed() bool {
	return time.Now().Before(b.DelayEnd)
}

func (b *cassandraBackend) delay(seconds int) {
	b.setDelay(time.Now().Add(time.Duration(seconds) * time.Second))
}

// cassandraBackend must be made with this function
func NewCassandraBackend(hosts []string) *cassandraBackend {
	log.Println("Initializing cassandra backend")
	log.Printf("Cassandra Hosts: %v\n", hosts)
	b := new(cassandraBackend)
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "catchall"
	b.Cluster = cluster
	b.createSession()
	return b
}

type aggregatedResponse struct {
	Count  int64
	Bounce bool
}

func (b *cassandraBackend) update() ([]HostEntry, error) {
	var results []HostEntry
	query := "SELECT host, lastseen from hosts"
	iter := b.Session.Query(query).Iter()
	if iter.NumRows() == 0 {
		log.Printf("No hosts found")
	}
	m := map[string]interface{}{}
	for iter.MapScan(m) {
		t := m["lastseen"].(time.Time)
		dt := t.Add(time.Duration(10) * time.Minute)
		now := time.Now()
		if now.After(dt) {
			continue
		}
		results = append(results, HostEntry{
			Host: m["host"].(string),
			Seen: m["lastseen"].(time.Time),
		})
		m = map[string]interface{}{}
	}
	return results, nil
}

type currentCounts struct {
	C      int
	Bounce bool
}

// Implements backendWriter
func (b *cassandraBackend) insert(entry DomainEntry) error {
	log.Printf("Inserting into backend %s\n", entry.String())
	domain := entry.Domain
	query := "SELECT c, bounce from events where domain = ?"
	iter := b.Session.Query(query, domain).Iter()
	var currentCount currentCounts = currentCounts{C: 0, Bounce: false}
	if iter.NumRows() > 0 {
		m := map[string]interface{}{}
		for iter.MapScan(m) {
			currentCount.C = m["c"].(int)
			currentCount.Bounce = m["bounce"].(bool)
			m = map[string]interface{}{}
		}
	}
	if err := b.Session.Query(
		`SELECT domain, c from events WHERE domain = ?`,
		entry.Domain).Exec(); err != nil {
		log.Printf("failed select: %s\n", err)
		return err
	}
	totalCount := entry.C + currentCount.C
	bouncy := entry.IsBounce || currentCount.Bounce
	if err := b.Session.Query(
		`INSERT INTO events (domain, bounce, c) VALUES (?, ?, ?)`,
		entry.Domain, bouncy, totalCount).Exec(); err != nil {
		log.Printf("failed insert: %s\n", err)
		return err
	}
	return nil
}

func (b *cassandraBackend) ping(entry HostEntry) error {
	if err := b.Session.Query(
		`INSERT INTO hosts (host, lastseen) VALUES (?, toTimestamp(now()))`,
		entry.Host).Exec(); err != nil {
		log.Fatalf("failed insert. %s\n", err)
	}
	return nil
}

// Implements backendReader
func (b *cassandraBackend) get(domain string) (QueryResponse, error) {
	var results []QueryResponse
	m := map[string]interface{}{}
	query := "SELECT domain, c, bounce from events where domain = ?"
	iter := b.Session.Query(query, domain).Iter()
	if iter.NumRows() == 0 {
		return QueryResponse{}, nil // zeroed response object is understood missing
	}
	for iter.MapScan(m) {
		results = append(results, QueryResponse{
			Bounced: m["bounce"].(bool),
			Domain:  m["domain"].(string),
			Total:   m["c"].(int),
		})
		m = map[string]interface{}{}
	}
	hasBounce := false
	var deliveries int = 0
	for _, row := range results {
		if row.Bounced {
			hasBounce = true
			break
		}
		deliveries = row.Total
	}
	ret := QueryResponse{
		Domain:  domain,
		Bounced: hasBounce,
		Total:   deliveries,
	}
	return ret, nil
}
