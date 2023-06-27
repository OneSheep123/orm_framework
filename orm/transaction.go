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

// Session db和session的抽象
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

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

// RollbackIfNotCommit 只需要尝试回滚，如果此时事务已经被提交，或者被回滚掉了，那么就会得到 sql.ErrTxDone 错误
func (t *Tx) RollbackIfNotCommit() error {
	err := t.tx.Rollback()
	if err == sql.ErrTxDone {
		return nil
	}
	return err
}
