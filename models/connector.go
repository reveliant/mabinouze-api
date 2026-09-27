package models

import (
	"context"
	"database/sql"
)

type Connector struct {
	DB		*sql.DB
	Context context.Context
}

func (c *Connector) Conn(db *sql.DB, ctx context.Context) {
	c.DB = db
	c.Context = ctx
}

func (c *Connector) Exec(query string, args ...any) (sql.Result, error) {
	return c.DB.ExecContext(c.Context, query, args...)
}

func (c *Connector) Query(query string, args ...any) (*sql.Rows, error) {
	return c.DB.QueryContext(c.Context, query, args...)
}

func (c *Connector) QueryRow(query string, args ...any) *sql.Row {
	return c.DB.QueryRowContext(c.Context, query, args...)
}

func (c *Connector) Begin() (*sql.Tx, error) {
	return c.DB.BeginTx(c.Context, nil)
}