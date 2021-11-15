package handler

import "context"
import "io"
import "github.com/vmihailenco/taskq/v3"
import "github.com/vmihailenco/taskq/v3/memqueue"

// Placeholder struct for eventual usage of a proper external message queue client
// Currently just represents a in-memory message queue handle struct
type MemMQ struct {}

// Creates a new MemMQ instance
func NewMQ() *MemMQ {
    return &MemMQ{}
}

// Helper func to create and wrap a worker func into a route callback
func (this *MemMQ) RegisterWorker(name string, fn RouteFunc) RouteFunc {
    queue := memqueue.NewQueue(&taskq.QueueOptions{
        Name: name,
    })
    task := taskq.RegisterTask(&taskq.TaskOptions{
        Name: name,
        Handler: fn,
    })
    return func(w io.Writer, vars map[string]string) error {
        return queue.Add(task.WithArgs(context.Background(), w, vars))
    }
}

// Placeholder func for an eventual MQ client
func (this *MemMQ) Listen() error {
    return nil
}

// Placeholder func for an eventual MQ client
func (this *MemMQ) Shutdown() {
}
