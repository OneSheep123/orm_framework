// Package reflect create by chencanhua in 2023/5/11
package reflect

import "reflect"

func IterateFunc(entity any) (map[string]FuncInfo, error) {
	tOf := reflect.TypeOf(entity)
	numMethod := tOf.NumMethod()
	res := map[string]FuncInfo{}
	for index := 0; index < numMethod; index++ {
		method := tOf.Method(index)
		// 核心是这里
		fn := method.Func
		in := fn.Type().NumIn()
		inputTypes := make([]reflect.Type, 0, in)
		inputValues := make([]reflect.Value, 0, in)

		inputTypes = append(inputTypes, reflect.TypeOf(entity))
		inputValues = append(inputValues, reflect.ValueOf(entity))
		for i := 1; i < in; i++ {
			fnInType := fn.Type().In(i)
			inputTypes = append(inputTypes, fnInType)
			inputValues = append(inputValues, reflect.Zero(fnInType))
		}

		numOut := fn.Type().NumOut()
		outTypes := make([]reflect.Type, 0, numOut)
		for i := 0; i < numOut; i++ {
			outTypes = append(outTypes, fn.Type().Out(i))
		}

		callRes := fn.Call(inputValues)
		result := make([]any, 0, len(callRes))
		for _, c := range callRes {
			result = append(result, c.Interface())
		}

		res[method.Name] = FuncInfo{
			Name:        method.Name,
			InputTypes:  inputTypes,
			OutputTypes: outTypes,
			Result:      result,
		}
	}
	return res, nil
}

type FuncInfo struct {
	Name string
	// 方法的输入类型
	InputTypes  []reflect.Type
	OutputTypes []reflect.Type
	Result      []any
}
