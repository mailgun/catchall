package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/tkanos/gonfig"
)

const (
	DOMAIN_KEY         = "domain"
	CONFIG             = "catchall.json"
	StatusUnknown      = "unknown"
	StatusCatchAll     = "catch-all"
	StatusNotCatchAll  = "not catch-all"
	CatchAllTrigger    = 1000
	HasBounced         = -1
	UnrecordedBounce   = -2
	RetryInsertTime    = 60 // seconds
	QueueStoreFilename = "queue.csv"
)

func serveHttp(
	ctx context.Context,
	wg *sync.WaitGroup, queue Queue,
	backendRead Backend, port int) {
	defer wg.Done()
	router := mux.NewRouter()
	// Write mode (PUTs)
	router.Handle(
		"/events/{"+DOMAIN_KEY+"}/delivered",
		HandleRequest(false, queue)).Methods("PUT")
	router.Handle(
		"/events/{"+DOMAIN_KEY+"}/bounced",
		HandleRequest(true, queue)).Methods("PUT")
	router.Handle(
		"/stats/{"+DOMAIN_KEY+"}",
		HandleStatsQuery(queue)).Methods("GET")
	router.Handle(
		"/domains/{"+DOMAIN_KEY+"}",
		HandleQuery(backendRead)).Methods("GET")
	router.PathPrefix("/").HandlerFunc(CatchAllHandler)

	// NOTE(roaet): Possible performance increase with valyala/fasthttp
	server := http.Server{
		Addr: ":" + strconv.Itoa(port), Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	select {
	case <-ctx.Done():
		server.Shutdown(ctx)
	}
	log.Println("HTTP Server shut down")
}

func doWork(
	ctx context.Context,
	wg *sync.WaitGroup, thisHost string, queue Queue,
	backend Backend) {
	defer wg.Done()
	workerWait := 1 * time.Millisecond
loop:
	select {
	case <-ctx.Done():
		break
	case <-time.After(workerWait):
		entry := HostEntry{Host: thisHost}
		backend.ping(entry)
		_, err := backend.update()
		if err != nil {
			log.Printf("Error updating host list: %v\n", err)
		}
		workerWait = 60 * time.Second
		goto loop
	}
}

type Configuration struct {
	Port               int
	Worker, API, Local bool
	APIMode, Host      string
	CassandraHosts     []string
	RabbitURI          string
}

func initializeConfig(filepath string, config *Configuration) error {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Could not load user information: %s\n", err)
	}
	homedir := usr.HomeDir
	err = gonfig.GetConf(homedir+"/"+CONFIG, config)
	if err != nil {
		log.Fatalf("Could not load config: %s: %s\n", CONFIG, err)
	}
	return nil
}

func main() {
	var port int
	var host string
	var wg sync.WaitGroup

	pprof.StartCPUProfile(os.Stdout)
	defer pprof.StopCPUProfile()

	configuration := Configuration{}
	initializeConfig(CONFIG, &configuration)

	flag.IntVar(&port, "port", configuration.Port, "port to listen on")
	flag.StringVar(&host, "host", configuration.Host, "host ip to listen on")
	flag.Parse()
	log.Printf("Host: %s\n", host)
	log.Printf("Port: %d\n", port)
	host_str := host + ":" + strconv.Itoa(port)
	hosts := []string{host_str}
	cass := NewCassandraBackend(configuration.CassandraHosts)
	defer cass.EndSession()
	httpClient := &DefaultHTTP{}
	intercom := NewHTTPIntercom(httpClient)
	backend, err := NewDistributedBackend(hosts, cass, intercom)
	if err != nil {
		cass.EndSession() // call manually because exit panics
		log.Fatalf("Error creating backend: %v\n", err)
	}
	queue := NewCachedQueue(cass)
	err = queue.InitializeQueueUnsafe()
	if err != nil {
		cass.EndSession() // call manually because exit panics
		log.Printf("Error initializing queue: %v\n", err)
	} else {
		log.Println("Queue initialized")
	}

	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go doWork(ctx, &wg, host_str, queue, backend)
	wg.Add(1)
	go serveHttp(ctx, &wg, queue, backend, port)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	cancel()
	wg.Wait()

	log.Println("Writing queue to file")
	err = queue.persist()
	if err != nil {
		log.Printf("Error persisting queue: %v\n", err)
	} else {
		log.Println("Queue saved")
	}
}
