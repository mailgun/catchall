package internal

import (
	"context"
	"database/sql"
	"fmt"

	catchall "github.com/ravlio/catch-all"

	// postgres driver, better to use pgx
	_ "github.com/lib/pq"
)

var _ catchall.Repository = (*Repository)(nil)

// Repository implements Repository interface
type Repository struct {
	dbs   []*sql.DB
	curDB int
}

// NewRepository creates new Repository instance
func NewRepository(dbs []*sql.DB) *Repository {
	r := Repository{
		dbs: dbs,
	}
	return &r
}

func (r *Repository) db() *sql.DB {
	r.curDB++
	return r.dbs[r.curDB%len(r.dbs)]
}

// Deliver atomically increment delivered counter for domain
func (r *Repository) Deliver(ctx context.Context, domain string) error {
	var ok int

	db := r.db()
	// try to increment counter via update
	err := db.QueryRowContext(
		ctx,
		queryDeliver,
		domain,
	).Scan(&ok)

	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("query failed (Deliver): %w", err)
		}

		// domain not found, so we just insert the new one
		// if at this time someone has already inserted same domain, we just update counter via on conflict do update
		_, err = db.ExecContext(context.Background(), queryDeliverInsert, domain)
		if err != nil {
			return fmt.Errorf("query failed (Deliver): %w", err)
		}
	}

	return nil
}

// Bounce atomically increment bounced counter for domain
func (r *Repository) Bounce(ctx context.Context, domain string) error {
	var ok int

	db := r.db()
	err := db.QueryRowContext(
		ctx,
		queryBounce,
		domain,
	).Scan(&ok)

	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("query failed (Bounce): %w", err)
		}

		_, err = db.ExecContext(context.Background(), queryBounceInsert, domain)
		if err != nil {
			return fmt.Errorf("query failed (Bounce): %w", err)
		}
	}

	return nil
}

// DomainStats collect delivered and bounced counters from db cluster
func (r *Repository) DomainStats(ctx context.Context, domain string) (*catchall.DomainStats, error) {
	ret := &catchall.DomainStats{
		Delivered: 0,
		Bounced:   0,
	}

	// iterate through all connections
	var found bool
	for _, db := range r.dbs {
		var delivered int
		var bounced int
		err := db.QueryRowContext(ctx, queryDomainStats, domain).Scan(&delivered, &bounced)

		if err != nil {
			if err != sql.ErrNoRows {
				return nil, fmt.Errorf("query failed (DomainStats): %w", err)
			}
		} else {
			// mark domain as found
			found = true
		}
		// increment total counters
		ret.Delivered += delivered
		ret.Bounced += bounced
	}

	// check if domain not found
	if !found {
		return nil, catchall.ErrDomainNotFound
	}

	return ret, nil
}
