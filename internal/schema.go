package internal

// Schema is table schema creation command.
const SchemaUp = `
CREATE TABLE IF NOT EXISTS domains
(
    domain    TEXT PRIMARY KEY,
    delivered INTEGER NOT NULL DEFAULT 0,
    bounced   INTEGER NOT NULL DEFAULT 0
)
`

// Down is table destruction command, used in tests.
const SchemaDown = `
drop table if exists domains;
`
