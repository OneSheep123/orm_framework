// create by chencanhua in 2023/5/16
package orm

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"orm_framework/orm/internal/errs"
	"testing"
)

func TestRegistry_get(t *testing.T) {
	type TestModel struct {
		Id        int
		FirstName string
		LastName  string
		Age       int8
	}
	testCases := []struct {
		name   string
		entity any

		wantRes   *Model
		wantError error
		cacheSize int
	}{
		{
			name:   "point struct",
			entity: &TestModel{},
			wantRes: &Model{
				tableName: "test_model",
				fieldMap: map[string]*field{
					"Id":        {colName: "id"},
					"FirstName": {colName: "first_name"},
					"LastName":  {colName: "last_name"},
					"Age":       {colName: "age"},
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
				tableName: "column_tag",
				fieldMap: map[string]*field{
					"ID": {
						colName: "id",
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
					FirstName uint64 `orm:"column="`
				}
				return &EmptyColumn{}
			}(),
			wantRes: &Model{
				tableName: "empty_column",
				fieldMap: map[string]*field{
					"FirstName": {
						colName: "first_name",
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
					FirstName uint64 `orm:"abc=abc"`
				}
				return &IgnoreTag{}
			}(),
			wantRes: &Model{
				tableName: "ignore_tag",
				fieldMap: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},

		{
			name:   "tableName  struct",
			entity: &User01{},
			wantRes: &Model{
				tableName: "user_01_t",
				fieldMap: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			name:   "tableName point  struct",
			entity: &User02{},
			wantRes: &Model{
				tableName: "user_02_t",
				fieldMap: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},

		{
			name:   "tableName empty",
			entity: &User03{},
			wantRes: &Model{
				tableName: "user03",
				fieldMap: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
	}

	r := newRegistry()
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
