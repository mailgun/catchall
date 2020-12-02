# catchall

A small p2p architecture that will allow the user to classify domains as catchall.

When a node (a single instance of catchall running) begins it will listen on the
provided host and port. It will also attempt to sync with the configured cassandra
backend. This sync includes its host information for other nodes to request on.
This attempt to join the 'cluster' of nodes does not have a particular 'main'
node, as all nodes are essentially the same.

When a node receives a PUT request on `../delivered` it will store it locally. When
the amount of delivered events reaches a (to be configured) value it will attempt
to persist that value on cassandra. It will also reset its internal counter to 0
so that there would not be a skew.
When a node receives a PUT request on `../bounced` it will immediately persist
that value on cassandra. It will also flag that domain as ignorable internally
so it can skip logic on it in the future.

A GET request on `../domains/..` will first attempt to gather information from
cassandra. If the information gathered does not determine the status of the domain
the node will gather more information from all known nodes (including itself) on
the `/stats/..` endpoint.  The nodes that are contacted must have performed
a sync (ping) within 10 minutes. 

(TODO) A node's worker will attempt to persist the values locally stored in 
memory to more durable storage should the process die.

## Requirements

- `cassandra`

## Creation

`go build`
`go install`

### Configuration

The application expects a JSON configuration file at `~/catchall.json`. Below is
an example of one:

```
{
  "Port": 8080,
  "Host": "127.0.0.1",
  "CassandraHosts": ["127.0.0.1"]
}
```

## Usage

This produces an executable `catchall` that accepts multiple arguments:

- `-port <n>` : the port used for the http server; default 8080
- `-host <ip>` : the host to use (interface?)

## Data Flow

The *write* flow is as follows:

`PUT X --> Local Cache`
 and after some time:
`Local Cache --> Backend`

The *read* flow is as follows:

`GET X --> Backend`
(if not already bounced):
`-> GET /stats/X on all nodes and events`

## Cassandra Tables

Events table to store persisted domain information

```
CREATE TABLE catchall.events (
    domain text PRIMARY KEY,
    bounce boolean,
    c int
)
```

Host table to store node information

```
CREATE TABLE catchall.hosts (
    host text PRIMARY KEY,
    lastseen timestamp
)
```

## Suggested optimizations

More nodes should provide increased throughput. For unified requests a load
balancer is recommended.

Some libraries may perform better than what is here:
- ristretto : a high performance and safe cache
- fasthttp : not a clean swap of http but has 10x performance

Other Suggestions:
- put a memory limit on the queue and evict on LRU basis to disk
- a reverse proxy can cache responses on GET for some period of time
- a static host for known non-catchall domains would be a very fast solution

## Future Recommended Changes

- auth support for servers
- coverage

## Resources

The following files are available for more information:

- `SPEC.md` the specifications of this project
- `NOTES.md` the notes of the developer
