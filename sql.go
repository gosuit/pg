package pg

import (
	"fmt"
	"reflect"
	"strings"
)

// TODO: error handling

type sqlFunc = func(model reflect.Value, args map[string]any) (sql string, sqlArgs []any)

func getSqlFunc(sql string, keys []valueKey, modelMeta *modelFields) sqlFunc {
	base := getSqlFuncBase(sql, keys, modelMeta)

	sqlFuncIn := []reflect.Type{reflect.TypeOf(reflect.Value{}), reflect.TypeOf(make(map[string]any))}
	sqlFuncOut := []reflect.Type{reflect.TypeOf(""), reflect.TypeOf([]any{})}
	sqlFuncType := reflect.FuncOf(sqlFuncIn, sqlFuncOut, false)

	return reflect.MakeFunc(sqlFuncType, base).Interface().(sqlFunc)
}

func getSqlFuncBase(sql string, keys []valueKey, modelMeta *modelFields) fnBase {
	return func(args []reflect.Value) (results []reflect.Value) {
		model := args[0].Interface().(reflect.Value)
		valueArgs := args[1].Interface().(map[string]any)

		sqlArgs := []any{}

		for _, k := range keys {
			if k.isModel {
				if model.Kind() != reflect.Struct {
					panic("can`t use array as src for sql")
				}
				
				sqlArgs = append(sqlArgs, modelMeta.getters[k.key](model).Interface())
			} else {
				sqlArgs = append(sqlArgs, valueArgs[k.key])
			}
		}

		results = append(results, reflect.ValueOf(sql))
		results = append(results, reflect.ValueOf(sqlArgs))

		return results
	}
}

type valueKey struct {
	key     string
	isModel bool
}

const ch = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func extractKeys(sql string) (string, []valueKey) {
	collectKey := false
	collectedKey := ""
	keys := []valueKey{}
	toReplace := []string{}

	for _, s := range sql {
		symb := string(s)

		if symb == "@" || symb == "#" {
			if collectKey {
				panic("invalid key")
			} else {
				collectKey = true
				collectedKey = symb
			}
		} else {
			if collectKey {
				if strings.Contains(ch, symb) {
					collectedKey += symb
				} else {
					if len(collectedKey) == 1 {
						panic("empty key")
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

	return sql, keys
}
