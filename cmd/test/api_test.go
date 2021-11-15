package test

import "fmt"
import "strings"
import "testing"
import "time"
import "cad/src/api"
import "cad/src/handler"
import "github.com/mailgun/catchall"

// Test the API methods
func TestAPI(t *testing.T) {
    state, err := api.NewAPIState()
    if err != nil {
        t.Fatal(err)
    }
    defer state.Clean()

    max_workers := 8
    task_queue := make(chan handler.RouteFunc, max_workers)
    arg_queue := make(chan map[string]string, max_workers)
    expected := map[string]int{}
    fn := map[string]handler.RouteFunc{
        "bounced": state.Bounced,
        "delivered": state.Delivered,
    }

    // Attempting to submit up to 10k events to the API event routes
    // while building an expected results map
    go func() {
        bus := catchall.SpawnEventPool()
        defer bus.Close()
        for i := 0; i < 10_000; i++ {
            e := bus.GetEvent()
            if e.Type == catchall.TypeBounced {
                expected[e.Domain] = -1
            } else if n, ok := expected[e.Domain]; !ok {
                expected[e.Domain] = 1
            } else if (n + 1) * (1000 - n) > 0 {
                expected[e.Domain] = n + 1
            }

            arg_queue <- map[string]string{"domain": e.Domain}
            task_queue <- fn[e.Type]
            bus.RecycleEvent(e)
        }
        close(arg_queue)
        close(task_queue)
    }()

    // Setting up a simple worker pool for parallel processing
    start := time.Now()
    t.Run("process events", func(t *testing.T) {
        for i := 0; i < max_workers; i++ {
            t.Run("", func(t *testing.T) {
                t.Parallel()
                for task := range task_queue {
                    err := task(nil, <-arg_queue)
                    if err != nil {
                        t.Error(err)
                    }
                }
            })
        }
    })
    duration := time.Since(start)
    fmt.Printf(
        "Processed by %d workers in %dms\n",
        max_workers,
        duration.Milliseconds(),
    )

    expected_string := func(i int) string {
        switch i {
        case 1000: return "catch-all"
        case -1: return "not catch-all"
        default: return "unknown"
        }
    }

    // Checking the API state against the expected result map
    t.Run("check results", func(t *testing.T) {
        for d, n := range expected {
            t.Run(d, func(t *testing.T) {
                t.Parallel()
    
                var b strings.Builder
                err := state.FetchDomain(&b, map[string]string{"domain": d})
                if err != nil {
                    t.Error(err)
                }

                x, y := expected_string(n), b.String()
                if x != y {
                    t.Errorf("expected '%s' (%d), got '%s'\n", x, n, y)
                }
            })
        }
    })
}
