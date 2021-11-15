package handler

import "context"
import "io"
import "log"
import "net/http"
import "github.com/gorilla/mux"

// Simple route helper types
type RouteFunc = func(w io.Writer, vars map[string]string) error
type Route struct {
    Method, Path, Name string
    Func RouteFunc
}

// Simple handler including a router and an event broker
type Handler struct {
    err chan error
    router *mux.Router
    mq *MemMQ
}

// Creates an handler instance with default arguments
func NewHandler() *Handler {
    return &Handler{
        err: make(chan error, 2),
        router: mux.NewRouter().StrictSlash(true),
        mq: NewMQ(),
    }
}

// Register a basic route with its handler func
func (this *Handler) RegisterRoute(r *Route) {
    this.router.
        Methods(r.Method).
        Path(r.Path).
        HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            err := r.Func(w, mux.Vars(req))
            if err != nil {
                log.Println(err)
            }
        })
}

// Register an event-backed route with its handler func
func (this *Handler) RegisterEvent(e *Route) {
    e.Func = this.mq.RegisterWorker(e.Name, e.Func)
    this.RegisterRoute(e)
}

// Start the server with gracefull shutdown
func (this *Handler) Listen() error {
    sig := NewSigInt()
    srv := http.Server{Addr: ":8080", Handler: this.router}

    go this.spawnMqFunc()
    go this.spawnSrvFunc(&srv)

    var err error
    select {
    case <-sig:
    case err = <-this.err:
        log.Println(err)
    }

    close(sig)
    srv.Shutdown(context.Background())
    this.mq.Shutdown()
    close(this.err)
    return err
}

func (this *Handler) spawnMqFunc() {
    err := this.mq.Listen()
    if err != nil {
        this.err <- err
    }
}

func (this *Handler) spawnSrvFunc(srv *http.Server) {
    err := srv.ListenAndServe()
    if err != http.ErrServerClosed {
        this.err <- err
    }
}
