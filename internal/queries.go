package internal

const queryDeliver = `UPDATE domains SET delivered=delivered + 1 WHERE domain = $1 RETURNING 1`
const queryBounce = `UPDATE domains SET bounced=bounced + 1 WHERE domain = $1 RETURNING 1`
const queryDeliverInsert = `INSERT INTO domains (domain, delivered, bounced)
	VALUES ($1, 1, 0)
	ON CONFLICT (domain) DO UPDATE SET delivered=domains.delivered + 1
`
const queryBounceInsert = `INSERT INTO domains (domain, delivered, bounced)
	VALUES ($1, 0, 1)
	ON CONFLICT (domain) DO UPDATE SET bounced=domains.bounced + 1
`
const queryDomainStats = `SELECT delivered, bounced FROM domains WHERE domain = $1`
