package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func outputJsonResponse(w http.ResponseWriter, body interface{}, status int) {
	jsonResponse, err := json.Marshal(body)
	if err != nil {
		status = http.StatusInternalServerError
		w.WriteHeader(status)
		log.Fatalf("Failure to marshal object: %v\n", body)
	}
	jsonResponse = append(jsonResponse, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}

func HandleStatsQuery(queue Queue) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		domain := vars[DOMAIN_KEY]
		stats := queue.peek(domain)
		s := StatsQuery{Domain: domain, Total: stats}
		status := http.StatusOK
		outputJsonResponse(w, s, status)
		log.Printf("%d:%s\n", status, s.String())
	})
}

func HandleRequest(bounce bool, queue Queue) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		// some amount of domain validation could be here
		entry := DomainEntry{
			Domain:   vars[DOMAIN_KEY],
			IsBounce: bounce,
		}
		status := http.StatusOK
		if err := queue.send(entry); err != nil {
			status = http.StatusInternalServerError
			log.Fatalf("Could not send to queue\n")
		}
		outputJsonResponse(w, entry, status)
	})
}

func HandleQuery(backend Backend) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		status := http.StatusOK
		domain := vars[DOMAIN_KEY]
		resp, err := backend.get(domain)
		if err != nil {
			status = http.StatusInternalServerError
			w.WriteHeader(status)
			log.Printf("Could not 'get' from backend: %s\n", err)
			return
		}
		if resp.Domain == "" {
			status = http.StatusNotFound
			log.Printf("%d:%s\n", status, domain)
			w.WriteHeader(status)
			return
		}
		s := StatusUnknown
		if resp.Bounced {
			s = StatusNotCatchAll
		} else if resp.Total >= CatchAllTrigger {
			s = StatusCatchAll
		}
		ds := DomainStatus{Domain: domain, Status: s}
		outputJsonResponse(w, ds, status)
		log.Printf("%d:%s\n", status, ds)
	})
}

func CatchAllHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
