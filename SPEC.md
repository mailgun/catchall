## Definition

catch-all domain: a domain name that ensure no mail to the domain is rejected or lost

  Ex: example.com
  Given mailboxes of the form X@example.com (dev@example.com, admin@example.com, etc)
  email sent to this address will never bounce or gt rejected

## Spec

### Functional requirements (FR)

1. Provide endpoint for PUT /events/<domain_name>/delivered
  a. anything not matching this pattern should 404

2. Provide endpoint for PUT /events/<domain_name>/bounced
  a. anything not matching this pattern should 404

3. Flag a domain if it has more than (>) 1000 delivered events as `catch-all`
  a. probably should be >=

4. Flag a domain if it has less than (<) 1000 delivered events as `unknown`

5. Flag a domain if it has ANY (!= 0) bounce events as `not catch-all`

6. Provide endpoint for GET /domains/<domain_name> that will return value of flag
   [`catch-all`, `unknown`, `not catch-all`]
  a. anything not matching this pattern should 404
  b. if domain not found 404
  c. if backend not available should 500

### Non-functional requirements (NR)

1. Distributed system that identifies catch-all domain by counting `delivered` and
   `bounced` events
  - Can be deployed on multiple nodes and support horizontal scaling

2. Fault tolerant architecture

3. 100,000 events per second across 10,000,000 different domain names with P95
   latency of 200ms
  - it does not need to immediately work at this scale just have an architecture
    that could support it

### General Requirements

- Written in go
- Any libraries or frameworks should be lightweight
- Use a DB from the following list: MongoDB, Cassandra, Postgres
- Submissions should be github PR and include @mkbond, @aerwin3, @thrawn01,
  @anton-efimenko
