// create by chencanhua in 2023/5/16
package model

import (
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"orm_framework/orm/internal/errs"
	"reflect"
	"testing"
)

func TestModelWithTableName(t *testing.T) {
	testCases := []struct {
		name          string
		val           any
		opt           ModelOpt
		wantTableName string
		wantErr       error
	}{
		{
			// 我们没有对空字符串进行校验
			name:          "empty string",
			val:           &TestModel{},
			opt:           WithTableName(""),
			wantTableName: "",
		},
		{
			name:          "table name",
			val:           &TestModel{},
			opt:           WithTableName("test_model_t"),
			wantTableName: "test_model_t",
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.val, tc.opt)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantTableName, m.TableName)
		})
	}
}

func TestModelWithColumnName(t *testing.T) {
	testCases := []struct {
		name        string
		val         any
		opt         ModelOpt
		field       string
		wantColName string
		wantErr     error
	}{
		{
			name:        "new name",
			val:         &TestModel{},
			opt:         WithColumnName("FirstName", "first_name_new"),
			field:       "FirstName",
			wantColName: "first_name_new",
		},
		{
			name:        "empty new name",
			val:         &TestModel{},
			opt:         WithColumnName("FirstName", ""),
			field:       "FirstName",
			wantColName: "",
		},
		{
			// 不存在的字段
			name:    "invalid Field name",
			val:     &TestModel{},
			opt:     WithColumnName("FirstNameXXX", "first_name"),
			field:   "FirstNameXXX",
			wantErr: errs.NewErrUnknownField("FirstNameXXX"),
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.val, tc.opt)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fd := m.FieldMap[tc.field]
			assert.Equal(t, tc.wantColName, fd.ColName)
		})
	}
}

func TestRegistry_get(t *testing.T) {
	type TestModel struct {
		Id        int64
		FirstName string
		Age       int8
		LastName  *sql.NullString
	}
	testCases := []struct {
		name   string
		entity any

		wantRes   *Model
		wantError error
		cacheSize int
	}{
		{
			name:      "test Model",
			entity:    TestModel{},
			wantError: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},
		{
			name:   "point struct",
			entity: &TestModel{},
			wantRes: &Model{
				TableName: "test_model",
				FieldMap: map[string]*Field{
					"Id": {
						ColName: "id",
						Type:    reflect.TypeOf(int64(0)),
						GoName:  "Id",
						Offset:  0,
					},
					"FirstName": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
						Offset:  8,
					},
					"Age": {
						ColName: "age",
						Type:    reflect.TypeOf(int8(0)),
						GoName:  "Age",
						Offset:  24,
					},
					"LastName": {
						ColName: "last_name",
						Type:    reflect.TypeOf(&sql.NullString{}),
						GoName:  "LastName",
						Offset:  32,
					},
				},
				ColumnMap: map[string]*Field{
					"id": {
						ColName: "id",
						Type:    reflect.TypeOf(int64(0)),
						GoName:  "Id",
						Offset:  0,
					},
					"first_name": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
						Offset:  8,
					},
					"age": {
						ColName: "age",
						Type:    reflect.TypeOf(int8(0)),
						GoName:  "Age",
						Offset:  24,
					},
					"last_name": {
						ColName: "last_name",
						Type:    reflect.TypeOf(&sql.NullString{}),
						GoName:  "LastName",
						Offset:  32,
					},
				},
			},
			wantError: nil,
			cacheSize: 1,
		},
		{
			// 多级指针
			name: "multiple pointer",
			// 因为 Go 编译器的原因，所以我们写成这样
			entity: func() any {
				val := &TestModel{}
				return &val
			}(),
			wantError: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},
		{
			name:      "map",
			entity:    map[string]string{},
			wantError: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},
		{
			name:      "slice",
			entity:    []int{},
			wantError: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},
		{
			name:      "basic type",
			entity:    0,
			wantError: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},

		// 标签相关测试用例
		{
			name: "column tag",
			entity: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type ColumnTag struct {
					ID uint64 `orm:"column=id"`
				}
				return &ColumnTag{}
			}(),
			wantRes: &Model{
				TableName: "column_tag",
				FieldMap: map[string]*Field{
					"ID": {
						ColName: "id",
						Type:    reflect.TypeOf(uint64(0)),
						GoName:  "ID",
					},
				},
				ColumnMap: map[string]*Field{
					"id": {
						ColName: "id",
						Type:    reflect.TypeOf(uint64(0)),
						GoName:  "ID",
					},
				},
			},
		},
		{
			// 如果用户设置了 column，但是传入一个空字符串，那么会用默认的名字
			name: "empty column",
			entity: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type EmptyColumn struct {
					FirstName string `orm:"column="`
				}
				return &EmptyColumn{}
			}(),
			wantRes: &Model{
				TableName: "empty_column",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
			},
		},
		{
			// 如果用户设置了 column，但是没有赋值
			name: "invalid tag",
			entity: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type InvalidTag struct {
					FirstName uint64 `orm:"column"`
				}
				return &InvalidTag{}
			}(),
			wantError: errs.NewErrInvalidTagContent("column"),
		},
		{
			// 如果用户设置了一些奇奇怪怪的内容，这部分内容我们会忽略掉
			name: "ignore tag",
			entity: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type IgnoreTag struct {
					FirstName string `orm:"abc=abc"`
				}
				return &IgnoreTag{}
			}(),
			wantRes: &Model{
				TableName: "ignore_tag",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
			},
		},

		{
			name:   "tableName  struct",
			entity: &User01{},
			wantRes: &Model{
				TableName: "user_01_t",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
			},
		},
		{
			name:   "tableName point  struct",
			entity: &User02{},
			wantRes: &Model{
				TableName: "user_02_t",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
			},
		},

		{
			name:   "tableName empty",
			entity: &User03{},
			wantRes: &Model{
				TableName: "user03",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						Type:    reflect.TypeOf(""),
						GoName:  "FirstName",
					},
				},
			},
		},
	}

	r := NewRegistry()
	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			m, err := r.Get(ts.entity)
			assert.Equal(t, ts.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, ts.wantRes, m)
		})
	}
}

type User01 struct {
	FirstName string `orm:"column=first_name"`
}

func (u User01) TableName() string {
	return "user_01_t"
}

type User02 struct {
	FirstName string `orm:"column=first_name"`
}

func (u *User02) TableName() string {
	return "user_02_t"
}

type User03 struct {
	FirstName string `orm:"column=first_name"`
}

func (u *User03) TableName() string {
	return ""
}

func Test_underscoreName(t *testing.T) {
	testCases := []struct {
		name    string
		srcStr  string
		wantStr string
	}{
		// 我们这些用例就是为了确保
		// 在忘记 underscoreName 的行为特性之后
		// 可以从这里找回来
		// 比如说过了一段时间之后
		// 忘记了 ID 不能转化为 id
		// 那么这个测试能帮我们确定 ID 只能转化为 i_d
		{
			name:    "upper cases",
			srcStr:  "ID",
			wantStr: "i_d",
		},
		{
			name:    "use number",
			srcStr:  "Table1Name",
			wantStr: "table1_name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := underscoreName(tc.srcStr)
			assert.Equal(t, tc.wantStr, res)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
