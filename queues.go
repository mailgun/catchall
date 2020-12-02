package main

import (
	"encoding/csv"
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
)

// A queue must be read and write
type Queue interface {
	send(DomainEntry) error
	pull(int) ([]DomainEntry, error)
	peek(string) int
	persist() error
}

type cachedQueue struct {
	store   map[string]int
	mu      sync.RWMutex
	backend Backend
}

func NewCachedQueue(persistanceBackend Backend) *cachedQueue {
	q := new(cachedQueue)
	q.store = make(map[string]int)
	q.backend = persistanceBackend
	return q
}

func (q *cachedQueue) statRead(domain string) (int, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	v, ok := q.store[domain]
	return v, ok
}

func (q *cachedQueue) statWrite(domain string, value int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.store[domain] = value
}

func (q *cachedQueue) send(entry DomainEntry) error {
	// stats is the state of the domain in the cache AND the count
	// of deliveries
	// stats >= 0									 : hasn't bounced, just a counter
	// stats == HasBounced				 : has bounced and been recorded
	// stats == UnrecordedBounce   : has bounced but not recorded
	var stats int
	if s, ok := q.statRead(entry.Domain); ok {
		stats = s
	}

	// counterValue is just so we can keep track of deliveries before bounce
	// this value would be seen as int(UnrecordedBounce) if it was that
	counterValue := stats
	if stats == HasBounced { // if bounced and recorded we ignore it
		return nil
	}

	// checking if unrecorded again to avoid incrementing
	if entry.IsBounce || stats == UnrecordedBounce {
		stats = UnrecordedBounce
	} else {
		stats++
		counterValue = stats // copying value to prevent weird logic later
	}

	doPersist := false
	// if the counter is high enough or it is unrecorded bounce
	// attempt to record (persist) it
	if stats >= CatchAllTrigger || stats == UnrecordedBounce {
		doPersist = true
	}

	// the way this delay works assumes that the domain will
	// come back up in the future; if it doesn't it will not be
	// persisted; a separate goroutine could do this concurrently
	// at the cost of some locking
	//
	// if the backend isDelayed then all inserts will be skipped
	// but the state will not change (or it will keep counting)
	if doPersist && !q.backend.isDelayed() {
		dom := DomainEntry{
			Domain:   entry.Domain,
			C:        counterValue,
			IsBounce: entry.IsBounce || stats == UnrecordedBounce,
		}
		err := q.backend.insert(dom)
		if err != nil {
			log.Printf("Error during backend insert. Delaying. %s\n", err)
			q.backend.delay(RetryInsertTime)
		} else {
			if stats == UnrecordedBounce {
				// bounce recorded and domain can be ignored locally
				stats = HasBounced
			} else {
				// counter value sent to backend, reset our local counter
				stats = 0
			}
		}
	}
	// update the cache with new state
	q.statWrite(entry.Domain, stats)
	return nil
}

func (q *cachedQueue) persistBounce(entry DomainEntry) error {
	if q.backend.isDelayed() {
		return errors.New("Backend is delayed when persisting bounce")
	}
	return nil
}

func (q *cachedQueue) persistDeliveryCounter(entry DomainEntry) error {
	if q.backend.isDelayed() {
		return errors.New("Backend is delayed when persisting counter")
	}
	err := q.backend.insert(entry)
	if err != nil {

	}
	return nil
}

func (q *cachedQueue) peek(domain string) int {
	stats, ok := q.statRead(domain)
	if !ok {
		return 0
	}
	return stats
}

func (q *cachedQueue) pull(batch int) ([]DomainEntry, error) {
	results := make([]DomainEntry, 0)
	return results, nil
}

func (q *cachedQueue) InitializeQueueUnsafe() error {
	file, err := os.Open(QueueStoreFilename)
	if err != nil {
		return err
	}
	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		log.Printf("Error initializing queue from file: %v\n", err)
		return err
	}
	var domain string
	var stats int64
	for _, row := range records {
		domain = row[0]
		if stats, err = strconv.ParseInt(row[1], 10, 64); err != nil {
			log.Printf("Error parsing int: %v\n", stats)
			continue
		}
		q.store[domain] = int(stats)
	}
	return nil
}

func (q *cachedQueue) convertQueueToCSVUnsafe() [][]string {
	numRecords := len(q.store)
	records := make([][]string, numRecords)
	i := 0
	for domain, stat := range q.store {
		records[i] = []string{domain, strconv.Itoa(stat)}
		i++
	}
	return records
}

func (q *cachedQueue) writeQueueToFileUnsafe() error {
	records := q.convertQueueToCSVUnsafe()

	// will truncate the file
	file, err := os.Create(QueueStoreFilename)
	if err != nil {
		return err
	}
	w := csv.NewWriter(file)
	defer w.Flush()
	for _, value := range records {
		err := w.Write(value)
		if err != nil {
			log.Printf("Error writing row %v\n", value)
			continue
		}
	}
	return nil
}

func (q *cachedQueue) persist() error {
	return q.writeQueueToFileUnsafe()
}
