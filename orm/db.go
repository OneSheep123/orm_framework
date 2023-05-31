// create by chencanhua in 2023/5/16
package orm

import "database/sql"

// DB 使用隔离的DB维护一个注册中心
type DB struct {
	r *registry
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
		r:  newRegistry(),
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func MustNewDB(driver string, dsn string, opts ...DBOptions) *DB {
	db, err := Open(driver, dsn, opts...)
	if err != nil {
		panic(err)
	}
	return db
}
