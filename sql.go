package pg

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type sqlFunc = func(model reflect.Value, args map[string]any) (sql string, sqlArgs []any, err error)

func getSqlFunc(sql string, keys []valueKey, modelMeta *modelFields) sqlFunc {
	base := getSqlFuncBase(sql, keys, modelMeta)

	sqlFuncIn := []reflect.Type{reflect.TypeFor[reflect.Value](), reflect.TypeFor[map[string]any]()}
	sqlFuncOut := []reflect.Type{reflect.TypeFor[string](), reflect.TypeFor[[]any](), reflect.TypeFor[error]()}
	sqlFuncType := reflect.FuncOf(sqlFuncIn, sqlFuncOut, false)

	return reflect.MakeFunc(sqlFuncType, base).Interface().(sqlFunc)
}

func getSqlFuncBase(sql string, keys []valueKey, modelMeta *modelFields) fnBase {
	return func(args []reflect.Value) (results []reflect.Value) {
		model := args[0].Interface().(reflect.Value)
		valueArgs := args[1].Interface().(map[string]any)

		sqlArgs := []any{}
		var err error = nil

		for _, k := range keys {
			if k.isModel {
				if model.Kind() != reflect.Struct {
					err = errors.New("can`t use array as src for sql")
				}

				getter, ok := modelMeta.getters[k.key]
				if ok {
					sqlArgs = append(sqlArgs, getter(model).Interface())
				} else {
					err = errors.New("model field not found")
					break
				}
			} else {
				value, ok := valueArgs[k.key]
				if ok {
					sqlArgs = append(sqlArgs, value)
				} else {
					err = errors.New("arg not found")
					break
				}
			}
		}

		results = append(results, reflect.ValueOf(sql))
		results = append(results, reflect.ValueOf(sqlArgs))

		if err != nil {
			results = append(results, reflect.ValueOf(err))
		} else {
			results = append(results, reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()))
		}

		return results
	}
}

type valueKey struct {
	key     string
	isModel bool
}

const ch = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func extractKeys(sql string) (string, []valueKey, error) {
	collectKey := false
	collectedKey := ""
	keys := []valueKey{}
	toReplace := []string{}

	for i, s := range sql {
		symb := string(s)

		if symb == "@" || symb == "#" {
			if collectKey {
				return "", nil, errors.New("invalid key")
			} else {
				collectKey = true
				collectedKey = symb
			}
		} else {
			if collectKey {
				if strings.Contains(ch, symb) {
					collectedKey += symb
				}

				if !strings.Contains(ch, symb) || i == len(sql)-1 {
					if len(collectedKey) == 1 {
						return "", nil, errors.New("empty key")
					}

					toReplace = append(toReplace, collectedKey)

					key := valueKey{
						key: collectedKey[1:],
					}

					if string(collectedKey[0]) == "@" {
						key.isModel = true
					} else {
						key.isModel = false
					}

					keys = append(keys, key)
					collectKey = false
				}
			}
		}
	}

	for i, k := range toReplace {
		sql = strings.ReplaceAll(sql, k, fmt.Sprintf("$%d", i+1))
	}

	return sql, keys, nil
}
