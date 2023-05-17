// create by chencanhua in 2023/5/16
package orm

import (
	"github.com/stretchr/testify/assert"
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

		wantRes   *model
		wantError error
		cacheSize int
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
			cacheSize: 1,
		},
	}

	r := newRegistry()
	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			m, err := r.get(ts.entity)
			assert.Equal(t, ts.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, ts.wantRes, m)
			assert.Equal(t, ts.cacheSize, len(r.models))
		})
	}
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
