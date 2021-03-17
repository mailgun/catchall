package internal

import (
	"context"
	"database/sql"
	"testing"

	catchall "github.com/ravlio/catch-all"
	"github.com/ravlio/catch-all/test"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	defer test.Setup(t, SchemaUp, SchemaDown)()
	db1 := test.Connect(t)
	defer db1.Close()

	repo := NewRepository([]*sql.DB{db1})
	t.Run("domain.com should return ErrDomainNotFound", func(t *testing.T) {
		_, err := repo.DomainStats(context.Background(), "domain.com")
		require.EqualError(t, err, catchall.ErrDomainNotFound.Error())
	})

	t.Run("deliver for domain.com", func(t *testing.T) {
		err := repo.Deliver(context.Background(), "domain.com")
		require.NoError(t, err)
	})

	t.Run("bounce for domain.com", func(t *testing.T) {
		err := repo.Bounce(context.Background(), "domain.com")
		require.NoError(t, err)
	})

	t.Run("deliver for domain.com again", func(t *testing.T) {
		err := repo.Deliver(context.Background(), "domain.com")
		require.NoError(t, err)
	})

	t.Run("bounce for domain.com again", func(t *testing.T) {
		err := repo.Bounce(context.Background(), "domain.com")
		require.NoError(t, err)
	})

	t.Run("get stats for domain.com", func(t *testing.T) {
		resp, err := repo.DomainStats(context.Background(), "domain.com")
		require.NoError(t, err)

		exp := &catchall.DomainStats{
			Delivered: 2,
			Bounced:   2,
		}
		require.Equal(t, exp, resp)
	})
}
