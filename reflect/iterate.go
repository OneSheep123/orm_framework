// Package reflect create by chencanhua in 2023/5/14
package reflect

import "reflect"

func IterateArrayOrSlice(entity any) ([]any, error) {
	vOf := reflect.ValueOf(entity)
	res := make([]any, 0, vOf.Len())
	for index := 0; index < vOf.Len(); index++ {
		res = append(res, vOf.Index(index).Interface())
	}
	return res, nil
}

func IterateMap(entity any) (keys []any, values []any, err error) {
	vOf := reflect.ValueOf(entity)
	v := vOf.MapRange()
	for v.Next() {
		keys = append(keys, v.Key().Interface())
		values = append(values, v.Value().Interface())
	}
	return
}
