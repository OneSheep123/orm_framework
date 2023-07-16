// create by chencanhua in 2023/6/26
package orm

import (
	"context"
	"orm_framework/orm/internal/valuer"
	"orm_framework/orm/model"
)

type core struct {
	valuer.Creator
	dialect Dialect
	// r 使用隔离的DB维护一个注册中心
	r     model.Registry
	model *model.Model
	mdls  []Middleware
}

func get[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	for index := len(c.mdls) - 1; index >= 0; index-- {
		root = c.mdls[index](root)
	}
	return root(ctx, qc)
}

func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	sql, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Result: nil,
			Err:    err,
		}
	}
	rows, err := sess.queryContext(ctx, sql.SQL, sql.Args...)
	// 注意这里查询完后要进行关闭，否则连接会无法释放
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return &QueryResult{
			Result: nil,
			Err:    err,
		}
	}
	if !rows.Next() {
		// 这里调用的是error下的ErrNoRows
		return &QueryResult{
			Result: nil,
			Err:    ErrNoRows,
		}
	}

	tp := new(T)
	meta, err := c.r.Get(tp)
	if err != nil {
		return &QueryResult{
			Result: nil,
			Err:    ErrNoRows,
		}
	}
	val := c.Creator(tp, meta)
	err = val.SetColumns(rows)
	return &QueryResult{
		Result: tp,
		Err:    err,
	}
}

func exec(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return execHandler(ctx, sess, qc)
	}
	for index := len(c.mdls) - 1; index >= 0; index-- {
		root = c.mdls[index](root)
	}
	return root(ctx, qc)
}

func execHandler(ctx context.Context, sess Session, qc *QueryContext) *QueryResult {
	query, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	result, err := sess.execContext(ctx, query.SQL, query.Args...)
	return &QueryResult{
		Result: result,
		Err:    err,
	}
}
