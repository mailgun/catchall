package internal

import (
	"context"
	"errors"
	"testing"

	catchall "github.com/ravlio/catch-all"
	"github.com/ravlio/catch-all/internal/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Deliver(t *testing.T) {
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

		err := svc.Deliver(context.Background(), "domain.com")
		require.NoError(t, err)
	})

	t.Run("deliver error", func(t *testing.T) {
		repo := &mock.Repository{
			DeliverFn: func(ctx context.Context, domain string) error {
				return errors.New("error")
			},
		}

		svc := NewService(WithRepository(repo))

		err := svc.Deliver(context.Background(), "domain.com")
		require.EqualError(t, err, "error")
	})
}

func TestService_Bounce(t *testing.T) {
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

		err := svc.Bounce(context.Background(), "domain.com")
		require.NoError(t, err)
	})

	t.Run("bounce error", func(t *testing.T) {
		repo := &mock.Repository{
			BounceFn: func(ctx context.Context, domain string) error {
				return errors.New("error")
			},
		}

		svc := NewService(WithRepository(repo))

		err := svc.Bounce(context.Background(), "domain.com")
		require.EqualError(t, err, "error")
	})
}

func TestService_DomainStats(t *testing.T) {
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

		status, err := svc.DomainStatus(context.Background(), "domain.com")
		require.NoError(t, err)
		require.Equal(t, catchall.DomainStatusCatchAll, status)
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

		status, err := svc.DomainStatus(context.Background(), "domain.com")
		require.NoError(t, err)
		require.Equal(t, catchall.DomainStatusBounced, status)
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

		status, err := svc.DomainStatus(context.Background(), "domain.com")
		require.NoError(t, err)
		require.Equal(t, catchall.DomainStatusUnknown, status)
	})

	t.Run("check error", func(t *testing.T) {
		repo := &mock.Repository{
			DomainStatusFn: func(ctx context.Context, domain string) (*catchall.DomainStats, error) {
				return nil, errors.New("error")
			},
		}

		svc := NewService(WithRepository(repo))

		_, err := svc.DomainStatus(context.Background(), "domain.com")
		require.EqualError(t, err, "error")
	})
}
