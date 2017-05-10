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
			name text NOT NULL, -- stores the name of the holiday
			holiday_date date NOT NULL, -- stores the date of the holiday
			active bool DEFAULT true,
			timestamp bigint DEFAULT EXTRACT(EPOCH FROM NOW())::bigint
		)
	*/
	table := schemamagic.NewTable(schemamagic.Table{Name: "sessions", DefaultSchema: defaultSchema, Database: database, Tx: tx})
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "id", Datatype: "bigserial", IsPrimary: true, IsUnique: true}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "name", Datatype: "text", IsNotNull: true}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "holiday_date", Datatype: "date", IsNotNull: true}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "active", Datatype: "boolean", DefaultExists: true, DefaultValue: "true"}))
	table.Append(schemamagic.NewColumn(schemamagic.Column{Name: "timestamp", Datatype: "bigint", DefaultExists: true, DefaultValue: "date_part('epoch'::text, now())::bigint"}))
	return table
}
