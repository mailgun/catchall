package internal

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	catchall "github.com/ravlio/catch-all"
	"github.com/ravlio/catch-all/internal/mock"
	"github.com/stretchr/testify/require"
)

func makeRequest(t *testing.T, method, addr string) (int, []byte) {
	u := url.URL{} // to prevent gosec
	ur, _ := u.Parse(addr)
	req, _ := http.NewRequestWithContext(context.Background(), method, ur.String(), nil)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("unexpected error on request")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("error on read body %s", err)
	}

	return resp.StatusCode, body
}

func TestHTTPTransport_Deliver(t *testing.T) {
	t.Run("deliver should be ok", func(t *testing.T) {
		repo := &mock.Repository{
			DeliverFn: func(ctx context.Context, domain string) error {
				if domain != "domain.com" {
					return errors.New("wrong domain")
				}
				return nil
			},
		}

		svc := NewService(WithRepository(repo))
		h := NewHTTPHandler(svc)
		srv := httptest.NewServer(h)
		defer srv.Close()

		status, _ := makeRequest(t, http.MethodPut, srv.URL+"/events/domain.com/delivered")
		require.Equal(t, http.StatusOK, status)
	})

	t.Run("deliver error", func(t *testing.T) {
		repo := &mock.Repository{
			DeliverFn: func(ctx context.Context, domain string) error {
				return errors.New("error")
			},
		}

		svc := NewService(WithRepository(repo))

		h := NewHTTPHandler(svc)
		srv := httptest.NewServer(h)
		defer srv.Close()

		status, _ := makeRequest(t, http.MethodPut, srv.URL+"/events/domain/delivered")
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

func TestHTTPTransport_Bounce(t *testing.T) {
	t.Run("bounce should be ok", func(t *testing.T) {
		repo := &mock.Repository{
			BounceFn: func(ctx context.Context, domain string) error {
				if domain != "domain.com" {
					return errors.New("wrong domain")
				}
				return nil
			},
		}

		svc := NewService(WithRepository(repo))
		h := NewHTTPHandler(svc)
		srv := httptest.NewServer(h)
		defer srv.Close()

		status, _ := makeRequest(t, http.MethodPut, srv.URL+"/events/domain.com/bounced")
		require.Equal(t, http.StatusOK, status)
	})

	t.Run("bounce error", func(t *testing.T) {
		repo := &mock.Repository{
			BounceFn: func(ctx context.Context, domain string) error {
				return errors.New("error")
			},
		}

		svc := NewService(WithRepository(repo))
		h := NewHTTPHandler(svc)
		srv := httptest.NewServer(h)
		defer srv.Close()

		status, _ := makeRequest(t, http.MethodPut, srv.URL+"/events/domain/bounced")
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

func TestHTTPTransport_DomainStats(t *testing.T) {
	t.Run("check DomainStatusCatchAll", func(t *testing.T) {
		repo := &mock.Repository{
			DomainStatusFn: func(ctx context.Context, domain string) (*catchall.DomainStats, error) {
				return &catchall.DomainStats{
					Delivered: catchall.CatchAllThreshold + 1,
					Bounced:   0,
				}, nil
			},
		}

		svc := NewService(WithRepository(repo))
		h := NewHTTPHandler(svc)
		srv := httptest.NewServer(h)
		defer srv.Close()

		status, resp := makeRequest(t, http.MethodGet, srv.URL+"/domains/domain")
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, `{"status":"catch-all"}`+"\n", string(resp))
	})

	t.Run("check DomainStatusBounced", func(t *testing.T) {
		repo := &mock.Repository{
			DomainStatusFn: func(ctx context.Context, domain string) (*catchall.DomainStats, error) {
				return &catchall.DomainStats{
					Delivered: catchall.CatchAllThreshold + 1,
					Bounced:   1,
				}, nil
			},
		}

		svc := NewService(WithRepository(repo))
		h := NewHTTPHandler(svc)
		srv := httptest.NewServer(h)
		defer srv.Close()

		status, resp := makeRequest(t, http.MethodGet, srv.URL+"/domains/domain")
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, `{"status":"bounced"}`+"\n", string(resp))
	})

	t.Run("check DomainStatusUnknown", func(t *testing.T) {
		repo := &mock.Repository{
			DomainStatusFn: func(ctx context.Context, domain string) (*catchall.DomainStats, error) {
				return &catchall.DomainStats{
					Delivered: catchall.CatchAllThreshold - 1,
					Bounced:   0,
				}, nil
			},
		}

		svc := NewService(WithRepository(repo))
		h := NewHTTPHandler(svc)
		srv := httptest.NewServer(h)
		defer srv.Close()

		status, resp := makeRequest(t, http.MethodGet, srv.URL+"/domains/domain")
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, `{"status":"unknown"}`+"\n", string(resp))
	})

	t.Run("check error", func(t *testing.T) {
		repo := &mock.Repository{
			DomainStatusFn: func(ctx context.Context, domain string) (*catchall.DomainStats, error) {
				return nil, errors.New("error")
			},
		}

		svc := NewService(WithRepository(repo))
		h := NewHTTPHandler(svc)
		srv := httptest.NewServer(h)
		defer srv.Close()
		status, _ := makeRequest(t, http.MethodGet, srv.URL+"/domains/domain")
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

func TestHTTPTransport_BadRequest(t *testing.T) {
	repo := &mock.Repository{
	}

	svc := NewService(WithRepository(repo))
	h := NewHTTPHandler(svc)
	srv := httptest.NewServer(h)
	defer srv.Close()

	status, resp := makeRequest(t, http.MethodGet, srv.URL+"/domains/do,main")
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, `{"error":"invalid domain","code":400}`+"\n", string(resp))
}
