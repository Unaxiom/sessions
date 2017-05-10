package sessions

import (
	"github.com/apratheek/schemamagic"
	pgx "gopkg.in/jackc/pgx.v2"
)

// TableSessions creates the Sessions table
func TableSessions(tx *pgx.Tx, defaultSchema string, database string) *schemamagic.Table {
	/*
		CREATE TABLE sessions (
			id bigserial UNIQUE PRIMARY KEY,
			key text NOT NULL, -- stores the session key
			value text UNIQUE NOT NULL, -- stores the generated session value, in response to the session key
			expires_at timestamp with timezone NOT NULL, -- stores the time when this session will expire
			ip text NOT NULL, -- stores the ip address of the client
			active bool DEFAULT true,
			timestamp bigint DEFAULT EXTRACT(EPOCH FROM NOW())::bigint
		)
	*/
	table := schemamagic.NewTable(schemamagic.Table{Name: "sessions", DefaultSchema: defaultSchema, Database: database, Tx: tx})
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "id", Datatype: "bigserial", IsPrimary: true, IsUnique: true}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "key", Datatype: "text", IsNotNull: true}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "value", Datatype: "text", IsNotNull: true, IsUnique: true}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "expires_at", Datatype: "timestamp", IsNotNull: true, PseudoDatatype: "timestamp with time zone"}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "ip", Datatype: "text", IsNotNull: true}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "active", Datatype: "boolean", DefaultExists: true, DefaultValue: "true"}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "timestamp", Datatype: "bigint", DefaultExists: true, DefaultValue: "date_part('epoch'::text, now())::bigint"}))
	return table
}
