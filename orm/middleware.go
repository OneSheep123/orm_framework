// create by chencanhua in 2023/6/29
package orm

import "context"

type QueryContext struct {
	Type    string
	Builder QueryBuilder
}

type QueryResult struct {
	Result any
	Err    error
}

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult

type Middleware func(next Handler) Handler
