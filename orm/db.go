// create by chencanhua in 2023/5/16
package orm

type DB struct {
	r *registry
}

type DBOptions func(db *DB)

func NewDB(opts ...DBOptions) (*DB, error) {
	db := &DB{
		r: newRegistry(),
	}
	for _, opt := range opts {
		opt(db)
	}
	return db, nil
}

func MustNewDB(opts ...DBOptions) *DB {
	db, err := NewDB(opts...)
	if err != nil {
		panic(err)
	}
	return db
}
