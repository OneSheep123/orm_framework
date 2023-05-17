// create by chencanhua in 2023/5/14
package orm

import (
	"github.com/stretchr/testify/assert"
	"orm_framework/orm/internal/errs"
	"testing"
)

func TestParseModel(t *testing.T) {
	type TestModel struct {
		Id        int
		FirstName string
		LastName  string
		Age       int8
	}
	testCases := []struct {
		name   string
		entity any

		wantRes   *model
		wantError error
	}{
		{
			name:   "point struct",
			entity: &TestModel{},
			wantRes: &model{
				tableName: "test_model",
				fieldMap: map[string]*field{
					"Id":        {colName: "id"},
					"FirstName": {colName: "first_name"},
					"LastName":  {colName: "last_name"},
					"Age":       {colName: "age"},
				},
			},
			wantError: nil,
		},
		{
			name:      "struct",
			entity:    TestModel{},
			wantError: errs.ErrPointOnly,
		},
	}

	r := newRegistry()
	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			m, err := r.parseModel(ts.entity)
			assert.Equal(t, ts.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, ts.wantRes, m)
		})
	}
}
