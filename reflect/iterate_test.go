// create by chencanhua in 2023/5/14
package reflect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIterateArrayOrSlice(t *testing.T) {
	testCases := []struct {
		name   string
		entity any

		wantRes   []any
		wantError error
	}{
		{
			name:      "array",
			entity:    [3]int{1, 2, 3},
			wantRes:   []any{1, 2, 3},
			wantError: nil,
		},
		{
			name:      "slice",
			entity:    []int{1, 2, 3},
			wantRes:   []any{1, 2, 3},
			wantError: nil,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			slice, err := IterateArrayOrSlice(ts.entity)
			assert.Equal(t, ts.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, ts.wantRes, slice)
		})
	}
}

func TestIterateMap(t *testing.T) {
	testCases := []struct {
		name   string
		entity any

		wantKeys   []any
		wantValues []any
		wantErr    error
	}{
		{
			name: "map",
			entity: map[string]string{
				"a": "A",
				"b": "B",
			},

			wantKeys:   []any{"a", "b"},
			wantValues: []any{"A", "B"},
			wantErr:    nil,
		},
	}
	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			keys, values, err := IterateMap(ts.entity)
			assert.Equal(t, ts.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, ts.wantValues, values)
			assert.Equal(t, ts.wantKeys, keys)
		})
	}
}
