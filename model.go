package pg

import (
	"errors"
	"maps"
	"reflect"
	"slices"
	"strings"
	"time"
)

const (
	pgTag = "pg"
)

var specificTypes = []reflect.Type{reflect.TypeFor[time.Time]()}

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
		meta.getters[k] = getGetter(v)
		meta.setters[k] = getSetter(v)
	}

	return meta, nil
}

func getPaths(modelType reflect.Type, baseKey string, basePath []int) map[string][]int {
	result := make(map[string][]int)

	for i := range modelType.NumField() {
		fieldStructType := modelType.Field(i)

		key, ok := fieldStructType.Tag.Lookup(pgTag)
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

		if fieldStructType.Type.Kind() != reflect.Struct || slices.Contains(specificTypes, fieldStructType.Type) {
			result[key] = path
		} else {
			toAdd := getPaths(modelType.Field(i).Type, key, path)

			maps.Copy(result, toAdd)
		}
	}

	return result
}

type setter = func(model reflect.Value, value reflect.Value) error
type getter = func(reflect.Value) reflect.Value

func getSetter(indexPath []int) setter {
	base := getSetterBase(indexPath)

	setterIn := []reflect.Type{reflect.TypeFor[reflect.Value](), reflect.TypeFor[reflect.Value]()}
	setterOut := []reflect.Type{reflect.TypeFor[error]()}
	setterType := reflect.FuncOf(setterIn, setterOut, false)

	return reflect.MakeFunc(setterType, base).Interface().(setter)
}

func getGetter(indexPath []int) getter {
	base := getGetterBase(indexPath)

	getterIn := []reflect.Type{reflect.TypeFor[reflect.Value]()}
	getterOut := []reflect.Type{reflect.TypeFor[reflect.Value]()}
	getterType := reflect.FuncOf(getterIn, getterOut, false)

	return reflect.MakeFunc(getterType, base).Interface().(getter)
}

func getSetterBase(indexPath []int) fnBase {
	return func(args []reflect.Value) (results []reflect.Value) {
		model := args[0].Interface().(reflect.Value)
		value := args[1].Interface().(reflect.Value)

		var field reflect.Value
		var err error

		for i := range indexPath {
			if i == 0 {
				field = model.Field(indexPath[i])
			} else {
				field = field.Field(indexPath[i])
			}
		}

		if !value.IsValid() {
			field.SetZero()
		} else if value.CanConvert(field.Type()) {
			field.Set(value.Convert(field.Type()))
		} else {
			err = errors.New("invalid value")
		}

		if err != nil {
			results = append(results, reflect.ValueOf(err))
		} else {
			results = append(results, reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()))
		}

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
