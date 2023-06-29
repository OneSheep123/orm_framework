// create by chencanhua in 2023/6/26
package orm

import (
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
