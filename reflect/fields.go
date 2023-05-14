// Package reflect create by chencanhua in 2023/5/9
package reflect

import (
	"errors"
	"reflect"
)

// IterateFields 遍历字段
func IterateFields(entity any) (map[string]any, error) {
	if entity == nil {
		return nil, errors.New("不支持 nil")
	}
	res := map[string]any{}
	tOf := reflect.TypeOf(entity)
	vOf := reflect.ValueOf(entity)
	if vOf.IsZero() {
		return nil, errors.New("不支持零值")
	}
	for tOf.Kind() == reflect.Ptr {
		tOf = tOf.Elem()
		vOf = vOf.Elem()
	}

	if tOf.Kind() != reflect.Struct {
		return nil, errors.New("不支持类型")
	}

	numField := tOf.NumField()
	for index := 0; index < numField; index++ {
		field := tOf.Field(index)
		if field.IsExported() {
			res[field.Name] = vOf.Field(index).Interface()
		} else {
			res[field.Name] = reflect.Zero(field.Type).Interface()
		}
	}
	return res, nil
}

func SetFields(entity any, field string, value any) error {
	val := reflect.ValueOf(entity)
	// val.Type().Kind() == reflect.Pointer
	for val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("不支持的类型")
	}
	fieldVal := val.FieldByName(field)
	if !fieldVal.CanSet() {
		return errors.New("不可修改字段")
	}
	fieldVal.Set(reflect.ValueOf(value))
	return nil
}

func SetMapField(entity any, field string, value any) (map[string]string, error) {
	vOf := reflect.ValueOf(entity)
	res := map[string]string{}
	for vOf.Kind() == reflect.Pointer {
		vOf = vOf.Elem()
	}
	if vOf.Kind() != reflect.Map {
		return nil, errors.New("不支持的类型")
	}
	mapRange := vOf.MapRange()
	for mapRange.Next() {
		if mapRange.Key().Interface() == field {
			vOf.SetMapIndex(mapRange.Key(), reflect.ValueOf(value))
		}
		res[mapRange.Key().Interface().(string)] = mapRange.Value().Interface().(string)
	}
	return res, nil
}
