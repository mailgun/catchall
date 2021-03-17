package internal

import (
	"context"

	"github.com/asaskevich/govalidator"
	catchall "github.com/ravlio/catch-all"
)

var _ catchall.Service = (*Service)(nil)

// Option is an option for Service
type Option func(*Service)

// WithRepository adds repository
func WithRepository(repository catchall.Repository) Option {
	return func(svc *Service) {
		svc.repo = repository
	}
}

func NewService(opts ...Option) *Service {
	svc := &Service{}
	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

// Service is implementation of catchall.Service interface
type Service struct {
	repo catchall.Repository
}

// Deliver increment delivered counter for domain
func (s *Service) Deliver(ctx context.Context, domain string) error {
	if !govalidator.IsHost(domain) {
		return catchall.ErrInvalidDomain
	}

	return s.repo.Deliver(ctx, domain)
}

// Bounce increment bounced counter for domain
func (s *Service) Bounce(ctx context.Context, domain string) error {
	if !govalidator.IsHost(domain) {
		return catchall.ErrInvalidDomain
	}

	return s.repo.Bounce(ctx, domain)
}

// DomainStatus retrieve domain status according to its delivered and bounced counters
func (s *Service) DomainStatus(ctx context.Context, domain string) (catchall.DomainStatus, error) {
	if !govalidator.IsHost(domain) {
		return "", catchall.ErrInvalidDomain
	}

	stats, err := s.repo.DomainStats(ctx, domain)
	if err != nil {
		return "", err
	}

	if stats.Bounced > 0 {
		return catchall.DomainStatusBounced, nil
	} else if stats.Delivered < catchall.CatchAllThreshold {
		return catchall.DomainStatusUnknown, nil
	} else {
		return catchall.DomainStatusCatchAll, nil
	}

}
