package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	catchall "github.com/ravlio/catch-all"
	"github.com/ravlio/catch-all/internal"
)

type stringSlice struct {
	Values []string
}

func (ss *stringSlice) String() string {
	return strings.Join(ss.Values, ",")
}
func (ss *stringSlice) Set(s string) error {
	ss.Values = append(ss.Values, s)
	return nil
}

func main() {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	apiAddr := fs.String("api.http-addr", ":8080", "HTTP ops API address to listen")
	pgDSNList := stringSlice{}
	fs.Var(&pgDSNList, "postgres.dsn-list", "Postgres dsn list")

	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "time", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var err error
	var repository catchall.Repository
	{
		var dbs = make([]*sql.DB, len(pgDSNList.Values))
		for k, dsn := range pgDSNList.Values {
			dbs[k], err = sql.Open("postgres", dsn)
			if err != nil {
				level.Error(logger).Log("message", "sql open failed", "err", err)
				os.Exit(1)
			}
			if err = dbs[k].Ping(); err != nil {
				level.Error(logger).Log("message", "sql ping failed", "err", err)
				os.Exit(1)
			}

			defer dbs[k].Close()
		}

		repository = internal.NewRepository(dbs)
	}

	svc := internal.NewService(internal.WithRepository(repository))

	handlers := internal.NewHTTPHandler(svc)
	apiServer := &http.Server{
		Addr:         *apiAddr,
		Handler:      handlers,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  70 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	var g run.Group
	{
		// prepare service to run
		g.Add(func() error {
			sig := make(chan os.Signal, 1)
			// wait for the signal
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-ctx.Done():
				return nil
			case s := <-sig:
				return fmt.Errorf("`%s` signal received", s.String())
			}
		}, func(err error) {
			_ = logger.Log("message", "program was interrupted", "err", err)
			cancel()
		})
	}
	{
		// prepare http api server
		g.Add(func() error {
			_ = logger.Log("message", "api server is starting", "addr", apiServer.Addr)
			return apiServer.ListenAndServe()
		}, func(err error) {
			_ = logger.Log("message", "api server was interrupted", "err", err)
			if err := apiServer.Shutdown(ctx); err != nil {
				_ = level.Error(logger).Log("message", "api server shut down", "err", err)
			}
		})
	}

	// run everything
	err = g.Run()
	_ = logger.Log("message", "actors stopped", "err", err)
}
