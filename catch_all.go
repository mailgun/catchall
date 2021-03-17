package catchall

import (
	"context"
	"errors"
)

// CatchAllThreshold is the minimal delivered events for the domain to be considered as a catch-all
const CatchAllThreshold = 10_000

// DomainStatus is the status of the domain
type DomainStatus string

// ErrDomainNotFound domain not found
var ErrDomainNotFound = errors.New("domain not found")

// ErrInvalidDomain invalid domain, bad request
var ErrInvalidDomain = errors.New("invalid domain")

const (
	// DomainStatusCatchAll for catch-all domain
	DomainStatusCatchAll DomainStatus = "catch-all"
	// DomainStatusBounced for bounced domain
	DomainStatusBounced DomainStatus = "bounced"
	// DomainStatusUnknown for unknown domain
	DomainStatusUnknown DomainStatus = "unknown"
)

// DomainStats stats response
type DomainStats struct {
	Delivered int
	Bounced   int
}

// Service is the main service interface
type Service interface {
	// Deliver increment delivered counter for domain
	Deliver(ctx context.Context, domain string) error
	// Bounce increment bounced counter for domain
	Bounce(ctx context.Context, domain string) error
	// DomainStatus retrieve domain status according to its delivered and bounced counters
	DomainStatus(ctx context.Context, domain string) (DomainStatus, error)
}

// Repository is the interface for database
type Repository interface {
	// Deliver atomically increment delivered counter for domain
	Deliver(ctx context.Context, domain string) error
	// Bounce atomically increment bounced counter for domain
	Bounce(ctx context.Context, domain string) error
	// DomainStats collect delivered and bounced counters from db cluster
	DomainStats(ctx context.Context, domain string) (*DomainStats, error)
}
