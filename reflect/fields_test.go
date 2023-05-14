// create by chencanhua in 2023/5/9
package reflect

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIterateFields(t *testing.T) {
	type User struct {
		Name string
		age  int
	}

	testCases := []struct {
		Name      string
		entity    any
		wantRes   any
		wantError error
	}{
		{
			Name:   "struct",
			entity: User{Name: "zhangsan", age: 18},
			wantRes: map[string]any{
				"Name": "zhangsan",
				"age":  0,
			},
			wantError: nil,
		},
		{
			Name:   "ptr struct",
			entity: &User{Name: "zhangsan", age: 18},
			wantRes: map[string]any{
				"Name": "zhangsan",
				"age":  0,
			},
			wantError: nil,
		},
		{
			Name: "multi ptr struct",
			entity: func() **User {
				res := &User{
					Name: "zhangsan",
					age:  18,
				}
				return &res
			}(),
			wantRes: map[string]any{
				"Name": "zhangsan",
				"age":  0,
			},
			wantError: nil,
		},
		{
			Name:      "base type",
			entity:    12,
			wantError: errors.New("不支持类型"),
		},
		{
			Name:      "nil",
			entity:    nil,
			wantError: errors.New("不支持 nil"),
		},
		{
			Name:      "user nil",
			entity:    (*User)(nil),
			wantError: errors.New("不支持零值"),
		},
	}

	for _, ts := range testCases {
		t.Run(ts.Name, func(t *testing.T) {
			fields, err := IterateFields(ts.entity)
			assert.Equal(t, err, ts.wantError)
			if err != nil {
				return
			}
			assert.Equal(t, fields, ts.wantRes)
		})
	}
}

func TestSetFields(t *testing.T) {
	type User struct {
		Name string
		age  int
	}
	testCases := []struct {
		name     string
		entity   any
		field    string
		newValue any

		wantEntity any
		wantError  error
	}{
		{
			name:     "struct",
			entity:   User{Name: "zhangsan"},
			field:    "Name",
			newValue: "wangwu",

			wantError: errors.New("不可修改字段"),
		},

		{
			name:       "ptr struct",
			entity:     &User{Name: "zhangsan"},
			field:      "Name",
			newValue:   "wangwu",
			wantEntity: &User{Name: "wangwu"},
			wantError:  nil,
		},

		{
			name: "map srt",
			entity: map[string]string{
				"name": "lisi",
			},
			field:     "name",
			newValue:  "wangwu",
			wantError: errors.New("不支持的类型"),
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			err := SetFields(ts.entity, ts.field, ts.newValue)
			assert.Equal(t, err, ts.wantError)
			if err != nil {
				return
			}
			assert.Equal(t, ts.entity, ts.wantEntity)
		})
	}
}

func TestSetMapField(t *testing.T) {
	testCases := []struct {
		name   string
		entity any

		field      string
		value      string
		wantEntity map[string]string
		wantError  error
	}{
		{
			name: "map",
			entity: map[string]string{
				"name": "lisi",
			},
			field: "name",
			value: "wangwu",

			wantEntity: map[string]string{
				"name": "wangwu",
			},
			wantError: nil,
		},

		{
			name: "point map",
			entity: &map[string]string{
				"name": "lisi",
			},
			field: "name",
			value: "wangwu",

			wantEntity: map[string]string{
				"name": "wangwu",
			},
			wantError: nil,
		},
	}
	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			res, err := SetMapField(ts.entity, ts.field, ts.value)
			assert.Equal(t, ts.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, ts.wantEntity, res)
		})
	}
}
