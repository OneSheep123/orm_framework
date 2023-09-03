// create by chencanhua in 2023/5/16
package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"orm_framework/orm/internal/errs"
	"orm_framework/orm/internal/valuer"
	"orm_framework/orm/model"
)

var (
	_ Session = &DB{}
)

type DB struct {
	core

	// db 使用到了装饰器模式
	db *sql.DB
}

type DBOptions func(db *DB)

func Open(driver string, dsn string, opts ...DBOptions) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

// OpenDB
// 用户可能自己创建了 sql.DB 实例，另外OpenDB一般也用于测试使用
func OpenDB(db *sql.DB, opts ...DBOptions) (*DB, error) {
	res := &DB{
		core: core{
			r:       model.NewRegistry(),
			Creator: valuer.NewUnsafeValue,
			dialect: MySQLDialect,
		},
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func WithMySQLDialect() DBOptions {
	return func(db *DB) {
		db.dialect = MySQLDialect
	}
}

func WithSqlite3Dialect() DBOptions {
	return func(db *DB) {
		db.dialect = SQLLiteDialect
	}
}

func WithDialect(dialect Dialect) DBOptions {
	return func(db *DB) {
		db.dialect = dialect
	}
}

func WithRegistry(r model.Registry) DBOptions {
	return func(db *DB) {
		db.r = r
	}
}

func WithReflectValue() DBOptions {
	return func(db *DB) {
		db.Creator = valuer.NewReflectValue
	}
}

func WithMiddleWare(mdls ...Middleware) DBOptions {
	return func(db *DB) {
		db.mdls = mdls
	}
}

func MustNewDB(driver string, dsn string, opts ...DBOptions) *DB {
	db, err := Open(driver, dsn, opts...)
	if err != nil {
		panic(err)
	}
	return db
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
}

func (db *DB) queryContext(context context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(context, query, args...)
}

func (db *DB) execContext(context context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(context, query, args...)
}

func (db *DB) getCore() core {
	return db.core
}

// DoTx 闭包事务
func (db *DB) DoTx(context context.Context,
	fn func(ctx context.Context, tx *Tx) error, sqlOptions *sql.TxOptions) (err error) {
	var tx *Tx
	tx, err = db.BeginTx(context, sqlOptions)
	if err != nil {
		return err
	}
	panicked := true

	defer func() {
		if !panicked || err != nil {
			e := tx.Rollback()
			err = errs.NewErrFailedToRollbackTx(err, e, panicked)
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(context, tx)
	panicked = false
	return err
}

func (db *DB) Wait() error {
	err := db.db.Ping()
	for err == driver.ErrBadConn {
		log.Println("数据库启动中")
		err = db.db.Ping()
	}
	return nil
}
