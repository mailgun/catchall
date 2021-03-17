# Catch-all service

- Clean architecture. API layer, Business Logic layer, Database layer.
- Horizontally scalable atomic PostgreSQL counters with optimistic upsert via `returning` syntax. Increments are routed between pg connections and then aggregated by stats request.
Just run multiple database instances as well as multiple service instances.
- Graceful start/stop with actors.
- Postgres-powered. I made atomic counters without slowdown of updates with additional checks.  
- Standard http server with Gorilla mux

## It would be good to add
- Aerospike/Couchbase instead of the Postgres. PG uses B-Tree for indexes with search complexity of log(log n), but for current task I would prefer hash tables for near O(1).
Also Aerospike has shored-nothing architecture and scalable out the box.
We also may use Cassandra, but not its default counters (https://ably.com/blog/cassandra-counter-columns-nice-in-theory-hazardous-in-practice). I would make this with Cassandra with simple GET/PUT operations with CAS (copy-and-swap) optimistic locking.
- https://github.com/valyala/fasthttp for http server instead of standard Golang http.
- write-ahead-log for increment operations. Realtime responses are not required, so we can put increments in WAL and periodically flush it into DB.
- Logging, Metrics, Instrumenting, Tracing.
