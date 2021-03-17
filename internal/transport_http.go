package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	catchall "github.com/ravlio/catch-all"
)

// NewHTTPHandler define http routes and handlers
func NewHTTPHandler(svc catchall.Service) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/events/{domain}/delivered", handleDelivered(svc)).Methods(http.MethodPut)
	r.HandleFunc("/events/{domain}/bounced", handleBounced(svc)).Methods(http.MethodPut)
	r.HandleFunc("/domains/{domain}", handleGetDomain(svc)).Methods(http.MethodGet)

	return r
}

func handleDelivered(svc catchall.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := svc.Deliver(r.Context(), mux.Vars(r)["domain"]); err != nil {
			errorHTTPResponse(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleBounced(svc catchall.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := svc.Bounce(r.Context(), mux.Vars(r)["domain"]); err != nil {
			errorHTTPResponse(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleGetDomain(svc catchall.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		status, err := svc.DomainStatus(r.Context(), mux.Vars(r)["domain"])
		if err != nil {
			errorHTTPResponse(w, err)
			return
		}

		resp := struct {
			Status string `json:"status"`
		}{
			Status: string(status),
		}

		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			errorHTTPResponse(w, err)
			return
		}

	}
}

func errorHTTPResponse(w http.ResponseWriter, err error) {
	var code int

	if errors.Is(err, catchall.ErrDomainNotFound) {
		code = http.StatusNotFound
	} else if errors.Is(err, catchall.ErrInvalidDomain) {
		code = http.StatusBadRequest
	} else {
		code = http.StatusInternalServerError
	}

	resp := struct {
		Error string `json:"error"`
		Code  int    `json:"code"`
	}{
		Error: err.Error(),
		Code:  code,
	}

	w.WriteHeader(code)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err.Error())
	}
}
