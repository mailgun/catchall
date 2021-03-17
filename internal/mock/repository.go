package mock

import "context"
import "github.com/ravlio/catch-all"

var _ catchall.Repository = (*Repository)(nil)

type Repository struct {
	DeliverFn      func(ctx context.Context, domain string) error
	BounceFn       func(ctx context.Context, domain string) error
	DomainStatusFn func(ctx context.Context, domain string) (*catchall.DomainStats, error)
}

func (r *Repository) Deliver(ctx context.Context, domain string) error {
	return r.DeliverFn(ctx, domain)
}

func (r *Repository) Bounce(ctx context.Context, domain string) error {
	return r.BounceFn(ctx, domain)
}

func (r *Repository) DomainStats(ctx context.Context, domain string) (*catchall.DomainStats, error) {
	return r.DomainStatusFn(ctx, domain)
}
