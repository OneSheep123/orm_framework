// create by chencanhua in 2023/5/14
package orm

import (
	"github.com/stretchr/testify/assert"
	"orm_framework/orm/internal/errs"
	"testing"
)

func TestDeleter_Build(t *testing.T) {
	type TestModel struct {
		Id        int
		FirstName string
		LastName  string
		Age       int8
	}
	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name:    "no where",
			builder: (&Deleter[TestModel]{}).From("`test_model`"),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name:    "where",
			builder: (&Deleter[TestModel]{}).Where(C("Id").Eq(16)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []any{16},
			},
		},
		{
			name:    "from",
			builder: (&Deleter[TestModel]{}).From("`test_model`").Where(C("Id").Eq(16)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []any{16},
			},
		},
		{
			name:    "from error",
			builder: (&Deleter[TestModel]{}).From("`test_model`").Where(C("XXX").Eq(16)),
			wantErr: errs.NewErrUnknownField("XXX"),
		},
	}

	for _, tc := range testCases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			query, err := c.builder.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}
