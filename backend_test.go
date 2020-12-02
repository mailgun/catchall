package main

import (
	"errors"
	"testing"
)

func TestCreateDistributedBackend(t *testing.T) {
	backend := NoopBackend{}
	httpClient := &MockedHTTP{}
	intercom := NewHTTPIntercom(httpClient)
	var tests = []struct {
		hostList []string
		backend  Backend
		isNil    bool
		isError  bool
		message  string
	}{
		{[]string{"host"}, backend, false, false,
			"Nonempty host list and non-nil backend should instantiate"},
		{[]string{"host"}, nil, true, true,
			"Nil host should error"},
		{[]string{}, backend, true, true,
			"Empty host list should fail"},
	}
	for _, test := range tests {
		res, err := NewDistributedBackend(test.hostList, test.backend, intercom)
		if test.isNil != (res == nil) {
			t.Errorf("(un)expected result to be nil %v\n", test)
		}
		if (err != nil) != test.isError {
			t.Errorf("(un)expected error: %s", test.message)
		}
	}
}

func TestDistributedBackendPing(t *testing.T) {
	httpClient := &MockedHTTP{}
	intercom := NewHTTPIntercom(httpClient)
	var tests = []struct {
		backend Backend
		isError bool
		message string
	}{
		{NoopBackend{FailPing: false}, false, "Ping should not fail"},
		{NoopBackend{FailPing: true}, true, "Ping should forward errors"},
	}
	hostList := []string{"host"}
	host := HostEntry{}
	for _, test := range tests {
		res, _ := NewDistributedBackend(hostList, test.backend, intercom)
		ret := res.ping(host)
		if ret != nil && !test.isError {
			t.Errorf("Ping should not return error with noop backend")
		}
	}
}

func TestDistributedBackendGet(t *testing.T) {
	var bounce interface{} = QueryResponse{
		Bounced: true}
	var nobounce interface{} = QueryResponse{Bounced: false}
	var added interface{} = QueryResponse{Total: 10}
	var tests = []struct {
		backend    Backend
		isError    bool
		domain     string
		message    string
		jsonBody   string
		status     int
		httpCalled bool
		delTotal   int
		bounced    bool
		httpErr    error
	}{ // future reference, use named params next time
		{NoopBackend{FailGet: false}, false, "widgetville.com",
			"Get should not fail in typical case",
			"{\"domain\":\"widgetville.com\", \"total\": 5}",
			200, true, 5, false, nil},
		{NoopBackend{FailGet: false}, false, "widgetville.com",
			"Get does not fail with bad json",
			"{\"domain\":\"widgetville.com\", total: 5}",
			200, true, 0, false, nil},
		{NoopBackend{FailGet: true}, false, "widgetville.com",
			"Get does not fail when backend fails",
			"{\"domain\":\"widgetville.com\", \"total\": 5}",
			200, true, 5, false, nil},
		{NoopBackend{FailGet: false, GetResp: bounce}, false, "widgetville.com",
			"Get should not http.get when domain has bounced from backend",
			"{\"domain\":\"widgetville.com\", \"total\": 5}",
			200, false, 0, true, nil},
		{NoopBackend{FailGet: false, GetResp: nobounce}, false, "widgetville.com",
			"Get should http.get when domain has not bounced from backend",
			"{\"domain\":\"widgetville.com\", \"total\": 5}",
			200, true, 5, false, nil},
		{NoopBackend{FailGet: false, GetResp: added}, false, "widgetville.com",
			"backend total should add with node totals",
			"{\"domain\":\"widgetville.com\", \"total\": 5}",
			200, true, 15, false, nil},
		{NoopBackend{FailGet: false, GetResp: added}, false, "widgetville.com",
			"if node reports < zero total, is bounce",
			"{\"domain\":\"widgetville.com\", \"total\": -1}",
			200, true, 10, true, nil},
		{NoopBackend{FailGet: false, GetResp: added}, false, "widgetville.com",
			"http get failure should not error",
			"{\"domain\":\"widgetville.com\", \"total\": 10}",
			200, true, 10, false, errors.New("simulated failure")},
		{NoopBackend{FailGet: true}, false, "",
			"Empty domain arg should do nothing",
			"{\"domain\":\"widgetville.com\", \"total\": 5}",
			200, false, 0, false, errors.New("SHOULD NOT SEE ME")},
	} // oh man use named params next time :(
	hostList := []string{"host"}
	for _, test := range tests {
		httpClient := &MockedHTTP{
			Ret:        test.jsonBody,
			StatusCode: test.status,
			Err:        test.httpErr}
		intercom := NewHTTPIntercom(httpClient)
		db, _ := NewDistributedBackend(hostList, test.backend, intercom)
		res, err := db.get(test.domain)
		if httpClient.Called != test.httpCalled {
			t.Errorf("(un)expected http call: %s", test.message)
		}
		if (err != nil) != test.isError {
			t.Errorf("(un)expected error: %s", test.message)
		}
		if res.Total != test.delTotal {
			t.Errorf("Expected (%d) totals do not match (got %d): %s",
				test.delTotal, res.Total, test.message)
		}
		if res.Bounced != test.bounced {
			t.Errorf("Expected (%v) bounce does not match (got %v): %s",
				test.bounced, res.Bounced, test.message)
		}
	}
}
