// create by chencanhua in 2023/5/16
package orm

import (
	"database/sql"
	"orm_framework/orm/internal/valuer"
	"orm_framework/orm/model"
)

type DB struct {
	// r 使用隔离的DB维护一个注册中心
	r model.Registry
	// db 使用到了装饰器模式
	db *sql.DB
	valuer.Creator
	dialect Dialect
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
		r:       model.NewRegistry(),
		db:      db,
		Creator: valuer.NewUnsafeValue,
		dialect: MySQLDialect,
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

func MustNewDB(driver string, dsn string, opts ...DBOptions) *DB {
	db, err := Open(driver, dsn, opts...)
	if err != nil {
		panic(err)
	}
	return db
}
