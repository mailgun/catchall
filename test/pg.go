package test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/lib/pq" // should be pgx
)

// Connect connect to database
func Connect(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := pq.NewConnector(os.Getenv("POSTGRES_DSN") + ";TimeZone=UTC")
	if err != nil {
		t.Fatal(err)
	}
	db := sql.OpenDB(conn)
	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

// Setup create all schema and return closure for teardown
func Setup(t *testing.T, up, down string) (teardown func()) {
	t.Helper()

	db := Connect(t)
	defer db.Close()
	if _, err := db.Exec(up); err != nil {
		t.Fatal(err)
	}

	return func() {
		t.Helper()
		db := Connect(t)
		defer db.Close()
		if _, err := db.Exec(down); err != nil {
			t.Fatal(err)
		}
	}
}
