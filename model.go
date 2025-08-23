package pg

import (
	"maps"
	"reflect"
	"strings"
)

// TODO: add support of specific types

type parsedModel struct {
	fields  *modelFields
	queries map[string]sqlFunc
}

type modelFields struct {
	getters map[string]getter
	setters map[string]setter
}

func parseModel(modelType reflect.Type) (*modelFields, error) {
	paths := getPaths(modelType, "", []int{})

	meta := &modelFields{
		getters: make(map[string]getter),
		setters: make(map[string]setter),
	}

	for k, v := range paths {
		meta.getters[k] = getGetter(v.indexPath)
		meta.setters[k] = getSetter(v.indexPath)
	}

	return meta, nil
}

type field struct {
	indexPath []int
}

func getPaths(modelType reflect.Type, baseKey string, basePath []int) map[string]field {
	result := make(map[string]field)

	for i := range modelType.NumField() {
		fieldStructType := modelType.Field(i)

		key, ok := fieldStructType.Tag.Lookup("pg")
		if ok && key == "-" {
			continue
		} else if !ok {
			key = strings.ToLower(fieldStructType.Name)
		}

		if baseKey != "" {
			key = baseKey + "." + key
		}

		path := make([]int, 0)

		if len(basePath) != 0 {
			path = append(basePath, i)
		} else {
			path = append(path, i)
		}

		if fieldStructType.Type.Kind() != reflect.Struct {
			f := field{
				indexPath: path,
			}

			result[key] = f
		} else {
			toAdd := getPaths(modelType.Field(i).Type, key, path)

			maps.Copy(result, toAdd)
		}
	}

	return result
}

type setter = func(model reflect.Value, value reflect.Value)
type getter = func(reflect.Value) reflect.Value

func getSetter(indexPath []int) setter {
	base := getSetterBase(indexPath)

	setterIn := []reflect.Type{reflect.TypeOf(reflect.Value{}), reflect.TypeOf(reflect.Value{})}
	setterType := reflect.FuncOf(setterIn, []reflect.Type{}, false)

	return reflect.MakeFunc(setterType, base).Interface().(setter)
}

func getGetter(indexPath []int) getter {
	base := getGetterBase(indexPath)

	getterIn := []reflect.Type{reflect.TypeOf(reflect.Value{})}
	getterOut := []reflect.Type{reflect.TypeOf(reflect.Value{})}
	getterType := reflect.FuncOf(getterIn, getterOut, false)

	return reflect.MakeFunc(getterType, base).Interface().(getter)
}

func getSetterBase(indexPath []int) fnBase {
	return func(args []reflect.Value) (results []reflect.Value) {
		model := args[0].Interface().(reflect.Value)
		value := args[1].Interface().(reflect.Value)

		var field reflect.Value

		for i := range indexPath {
			if i == 0 {
				field = model.Field(indexPath[i])
			} else {
				field = field.Field(indexPath[i])
			}
		}

		field.Set(value.Convert(field.Type()))

		return results
	}
}

func getGetterBase(indexPath []int) fnBase {
	return func(args []reflect.Value) (results []reflect.Value) {
		model := args[0].Interface().(reflect.Value)

		var result reflect.Value

		for i := range indexPath {
			if i == 0 {
				result = model.Field(indexPath[i])
			} else {
				result = result.Field(indexPath[i])
			}
		}

		results = append(results, reflect.ValueOf(result))

		return results
	}
}
