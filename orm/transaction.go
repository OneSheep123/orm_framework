// Package orm create by chencanhua in 2023/6/26
package orm

import (
	"context"
	"database/sql"
)

var (
	_ Session = &Tx{}
	_ Session = &DB{}
)

type Session interface {
	getCore() core
	queryContext(context context.Context, query string, args ...any) (*sql.Rows, error)
	execContext(context context.Context, query string, args ...any) (sql.Result, error)
}

type Tx struct {
	tx *sql.Tx
	db *DB
}

func (t *Tx) getCore() core {
	return t.db.core
}

func (t *Tx) queryContext(context context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(context, query, args)
}

func (t *Tx) execContext(context context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(context, query, args)
}
