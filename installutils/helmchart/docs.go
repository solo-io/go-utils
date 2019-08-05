package helmchart

import (
	"fmt"
	"reflect"
	"strings"
)

type HelmValue struct {
	Key          string
	Type         string
	DefaultValue string
	Description  string
}

type HelmValues []HelmValue

func (v HelmValues) ToMarkdown() string {
	result := new(strings.Builder)
	fmt.Fprintln(result, "|Option|Type|Default Value|Description|")
	fmt.Fprintln(result, "|------|----|-----------|-------------|")
	for _, value := range v {
		fmt.Fprintf(result, "|%s|%s|%s|%s|\n", value.Key, value.Type, value.DefaultValue, value.Description)
	}
	return result.String()
}

type addValue func(HelmValue)

func Doc(s interface{}) HelmValues {
	var values []HelmValue
	cfgT := reflect.ValueOf(s)
	addValue := func(v HelmValue) { values = append(values, v) }
	docReflect(addValue, nil, "", cfgT.Type(), cfgT)

	return HelmValues(values)
}

func docReflect(addValue addValue, path []string, desc string, typ reflect.Type, val reflect.Value) {
	switch typ.Kind() {
	case reflect.Ptr:
		var elemVal reflect.Value
		if elemVal != val {
			elemVal = val.Elem()
		}
		docReflect(addValue, path, desc, typ.Elem(), elemVal)
	case reflect.Map:
		if typ.Key().Kind() == reflect.String {
			docReflect(addValue, append(path, "NAME"), desc, typ.Elem(), reflect.Value{})

			if (val != reflect.Value{}) {
				for _, k := range val.MapKeys() {
					path = append(path, k.String())
					defaultVal := val.MapIndex(k)
					if typ.Elem().Kind() <= reflect.Float64 || typ.Elem().Kind() == reflect.String {
						// primitve type, print it as default value
						valStr := valToString(defaultVal)
						addValue(HelmValue{Key: strings.Join(path, "."), Type: typ.Elem().Kind().String(), DefaultValue: valStr, Description: desc})
					} else {
						// non primitive type, decend
						docReflect(addValue, path, desc, typ.Elem(), val.MapIndex(k))
					}
				}
			}
		}
	case reflect.Slice:
		lst := len(path) - 1
		path[lst] = path[lst] + "[]"
		docReflect(addValue, path, desc, typ.Elem(), reflect.Value{})
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			jsonTag := field.Tag.Get("json")
			parts := strings.Split(jsonTag, ",")
			jsonName := parts[0]
			desc := field.Tag.Get("desc")
			fieldPath := path
			if jsonName != "" {
				fieldPath = append(fieldPath, jsonName)
			}
			var fieldVal reflect.Value
			if val != fieldVal {
				fieldVal = val.Field(i)
			}
			docReflect(addValue, fieldPath, desc, field.Type, fieldVal)
		}
	default:
		addValue(HelmValue{Key: strings.Join(path, "."), Type: typ.Kind().String(), DefaultValue: valToString(val), Description: desc})
	}
}

func valToString(val reflect.Value) string {
	valStr := ""
	if val.IsValid() {
		switch val.Kind() {
		case reflect.Bool:
			valStr = fmt.Sprint(val.Bool())
		case reflect.String:
			valStr = fmt.Sprint(val.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valStr = fmt.Sprint(val.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			valStr = fmt.Sprint(val.Uint())
		}
	}
	return valStr
}
