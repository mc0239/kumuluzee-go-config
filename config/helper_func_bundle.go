package config

import (
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

func traverseStruct(s interface{}, prefixKey string, fieldProcessFunc func(key string, value reflect.Value, field reflect.StructField, tags reflect.StructTag)) {
	// passed value is not of type reflect.Value?
	// I will make passed value of type reflect.Value
	var val reflect.Value
	var valType reflect.Type
	val, ok := s.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(s).Elem()
	}
	valType = val.Type()

	// iterate through fields (assuming passed value was struct pointer!)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		//fieldName := valType.Field(i).Name
		fieldTags := valType.Field(i).Tag

		key := retrieveKey(prefixKey, valType.Field(i), fieldTags)
		// if field is a struct, recursively call function to traverse all nested structs aswell
		if field.Kind() == reflect.Struct {
			traverseStruct(field, key, fieldProcessFunc)
		} else {
			// field processing is only done on fields that aren't nested structs
			if fieldProcessFunc != nil {
				fieldProcessFunc(key, field, valType.Field(i), fieldTags)
			}
		}
	}
}

func retrieveKey(prefixKey string, field reflect.StructField, tags reflect.StructTag) string {
	// building key: if config tag is defined and has non-empty first value,
	// use prefixKey + tag, otherwise, use prefixKey + lowercased field name
	var key string

	if tag, ok := tags.Lookup("config"); ok {
		tvs := strings.Split(tag, ",")
		if tvs[0] != "" {
			key = prefixKey + "." + tvs[0]
		}
	} else {
		r, n := utf8.DecodeRuneInString(field.Name)
		lkey := string(unicode.ToLower(r)) + field.Name[n:]
		key = prefixKey + "." + lkey
	}

	return key
}

func setValueWithReflect(key string, value reflect.Value, field reflect.StructField, bun Bundle) {
	switch field.Type.Kind() {
	case reflect.Bool:
		if val, ok := bun.conf.GetBool(key); ok {
			value.Set(reflect.ValueOf(val))
		}
		break
	case reflect.String:
		if val, ok := bun.conf.GetString(key); ok {
			value.Set(reflect.ValueOf(val))
		}
		break
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		if val, ok := bun.conf.GetInt(key); ok {
			value.Set(reflect.ValueOf(val))
		}
		break
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		if val, ok := bun.conf.GetFloat(key); ok {
			value.Set(reflect.ValueOf(val))
		}
		break
	default:
		bun.Logger.Warning("Field %s could not be properly reflected, ignoring.", key)
		break
	}
}
