To avoid preoptimization I am breaking the project into MVPs.

## MVP 1

Support FR[1, 2, 6]

At this point will only return 404s and 500s.

- using net/http but valyala/fasthttp has 10x better perfomance

## MVP 2

Support everything else :P


## Benchmarks

Specs:

- single nodes
- VMware fusion ubuntu vm
- single node cassandra cluster on same machine
- nginx on same machine

Conclusions:

- The growth appears to be linear, as 10k is 10x faster than 100k consistently
- Adding nginx adds ~ 300k ns/op for no gain on a single machine
- Almost certain a single node cassandra cluster isn't optimal
- Adding additional nodes increases contention and slows it down by 50k ns/op
- Would love to see if lots of nodes and a beefier cassandra would improve it 

### 10000x no nginx (single node)

```
$ go test -bench=. -benchtime 10000x
goos: linux
goarch: amd64
pkg: github.com/mailgun/catchall
BenchmarkEventBus/GetEvent-4               10000            418812 ns/op
PASS
ok      github.com/mailgun/catchall     4.213s
```

### 10000x nginx (single node)

```
$ go test -bench=. -benchtime 10000x
goos: linux
goarch: amd64
pkg: github.com/mailgun/catchall
BenchmarkEventBus/GetEvent-4               10000            705743 ns/op
PASS
ok      github.com/mailgun/catchall     7.101s
```

### 100000x no nginx (single node)

```
$ go test -bench=. -benchtime 100000x
goos: linux
goarch: amd64
pkg: github.com/mailgun/catchall
BenchmarkEventBus/GetEvent-4              100000            414353 ns/op
PASS
ok      github.com/mailgun/catchall     41.472s
```

### 100000x nginx (single node)

```
$ go test -bench=. -benchtime 100000x
goos: linux
goarch: amd64
pkg: github.com/mailgun/catchall
BenchmarkEventBus/GetEvent-4              100000            700832 ns/op
PASS
ok      github.com/mailgun/catchall     70.111s
```

