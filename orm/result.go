// create by chencanhua in 2023/6/24
package orm

import "database/sql"

type Result struct {
	err error
	res sql.Result
}

func (r *Result) LastInsertId() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.LastInsertId()
}

func (r *Result) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.RowsAffected()
}
