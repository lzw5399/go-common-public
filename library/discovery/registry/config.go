package registry

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
)

var filedIndexes = map[string]int{}
var originRemoteKeys = map[string]bool{}

type ParserFunc func(v string) (interface{}, error)

type RemoteConfServer interface {
	GetAllKeys() []string
	Get(key string) string
	IsKeyLower() bool
}

var (
	ErrNotAStructPtr = errors.New("env: expected a pointer to a Struct")

	defaultBuiltInParsers = map[reflect.Kind]ParserFunc{
		reflect.Bool: func(v string) (interface{}, error) {
			return strconv.ParseBool(v)
		},
		reflect.String: func(v string) (interface{}, error) {
			return v, nil
		},
		reflect.Int: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int(i), err
		},
		reflect.Int16: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 16)
			return int16(i), err
		},
		reflect.Int32: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int32(i), err
		},
		reflect.Int64: func(v string) (interface{}, error) {
			return strconv.ParseInt(v, 10, 64)
		},
		reflect.Int8: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 8)
			return int8(i), err
		},
		reflect.Uint: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint(i), err
		},
		reflect.Uint16: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 16)
			return uint16(i), err
		},
		reflect.Uint32: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint32(i), err
		},
		reflect.Uint64: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 64)
			return i, err
		},
		reflect.Uint8: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 8)
			return uint8(i), err
		},
		reflect.Float64: func(v string) (interface{}, error) {
			return strconv.ParseFloat(v, 64)
		},
		reflect.Float32: func(v string) (interface{}, error) {
			f, err := strconv.ParseFloat(v, 32)
			return float32(f), err
		},
	}
)

// applyRemoteConfig
func applyRemoteConfig(config interface{}, remoteServer RemoteConfServer) error {
	ptrRef := reflect.ValueOf(config)
	if ptrRef.Kind() != reflect.Ptr {
		return ErrNotAStructPtr
	}
	ref := ptrRef.Elem()
	if ref.Kind() != reflect.Struct {
		return ErrNotAStructPtr
	}
	// 清空
	originRemoteKeys = make(map[string]bool)
	remoteKeys := remoteServer.GetAllKeys()
	for i := 0; i < len(remoteKeys); i++ {
		key := remoteKeys[i]
		originRemoteKeys[key] = true
	}
	var refType = ref.Type()
	for i := 0; i < refType.NumField(); i++ {
		refField := ref.Field(i)
		if !refField.CanSet() {
			continue
		}
		if reflect.Ptr == refField.Kind() && !refField.IsNil() {
			continue
		}
		if reflect.Struct == refField.Kind() && refField.CanAddr() && refField.Type().Name() == "" {
			continue
		}
		refTypeField := refType.Field(i)
		// 从环境变量的env tag中取出key, 一定是大写
		key, err := getKey(refTypeField)
		if err != nil {
			return err
		}
		// 记录字段的index
		filedIndexes[key] = i
		var remoteKey string
		if remoteServer.IsKeyLower() {
			remoteKey = strings.ToLower(key)
		} else {
			remoteKey = key
		}

		// 只初始化apollo中已设置的参数, 其他的忽略
		if _, ok := originRemoteKeys[remoteKey]; ok {
			value := remoteServer.Get(remoteKey)
			if err := set(refField, refTypeField, value); err != nil {
				return err
			}
			fmt.Printf("[remote-conf] set key [%v] value [%v]\n", remoteKey, value)
		}
	}
	return nil
}

func getKey(field reflect.StructField) (key string, err error) {
	key, opts := parseKeyForOption(field.Tag.Get("env"))

	for _, opt := range opts {
		switch opt {
		case "":
			break
		default:
			return "", fmt.Errorf("env: tag option %q not supported", opt)
		}
	}

	return strings.ToUpper(key), err
}

func parseKeyForOption(key string) (string, []string) {
	opts := strings.Split(key, ",")
	return opts[0], opts[1:]
}

func set(field reflect.Value, sf reflect.StructField, value string) error {
	if field.Kind() == reflect.Slice {
		return handleSlice(field, value, sf)
	}

	var typee = sf.Type
	var fieldee = field
	if typee.Kind() == reflect.Ptr {
		typee = typee.Elem()
		fieldee = field.Elem()
	}

	parserFunc, ok := defaultBuiltInParsers[typee.Kind()]
	if ok {
		val, err := parserFunc(value)
		if err != nil {
			return fmt.Errorf(`env: parse error on field "%s" of type "%s": %v`, sf.Name, sf.Type, err)
		}

		fieldee.Set(reflect.ValueOf(val))
		return nil
	}

	parserFunc, ok = defaultBuiltInParsers[typee.Kind()]
	if ok {
		val, err := parserFunc(value)
		if err != nil {
			return fmt.Errorf(`env: parse error on field "%s" of type "%s": %v`, sf.Name, sf.Type, err)
		}

		fieldee.Set(reflect.ValueOf(val).Convert(typee))
		return nil
	}

	return fmt.Errorf(`env: no parser found for field "%s" of type "%s"`, sf.Name, sf.Type)
}

func update(config interface{}, key, value string) error {
	ptrRef := reflect.ValueOf(config)
	if ptrRef.Kind() != reflect.Ptr {
		return ErrNotAStructPtr
	}
	ref := ptrRef.Elem()
	if ref.Kind() != reflect.Struct {
		return ErrNotAStructPtr
	}
	var refType = ref.Type()
	filedIndex, ok := filedIndexes[key]
	// 只处理已知的配置值
	if ok {
		refField := ref.Field(filedIndex)
		refTypeField := refType.Field(filedIndex)
		if err := set(refField, refTypeField, value); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return fmt.Errorf("handle OnUpdate:[%v]:[%v] ignore, unrecognize config", key, value)
	}
}

func handleSlice(field reflect.Value, value string, sf reflect.StructField) error {
	var separator = sf.Tag.Get("envSeparator")
	if separator == "" {
		separator = ","
	}
	var parts = strings.Split(value, separator)

	var typee = sf.Type.Elem()
	if typee.Kind() == reflect.Ptr {
		typee = typee.Elem()
	}

	if _, ok := reflect.New(typee).Interface().(encoding.TextUnmarshaler); ok {
		return parseTextUnmarshalers(field, parts, sf)
	}

	parserFunc, ok := defaultBuiltInParsers[typee.Kind()]
	if !ok {
		return fmt.Errorf(`env: no parser found for field "%s" of type "%s"`, sf.Name, sf.Type)
	}

	var result = reflect.MakeSlice(sf.Type, 0, len(parts))
	for _, part := range parts {
		r, err := parserFunc(part)
		if err != nil {
			return fmt.Errorf(`env: parse error on field "%s" of type "%s": %v`, sf.Name, sf.Type, err)
		}
		var v = reflect.ValueOf(r).Convert(typee)
		if sf.Type.Elem().Kind() == reflect.Ptr {
			v = reflect.New(typee)
			v.Elem().Set(reflect.ValueOf(r).Convert(typee))
		}
		result = reflect.Append(result, v)
	}
	field.Set(result)
	return nil
}

func parseTextUnmarshalers(field reflect.Value, data []string, sf reflect.StructField) error {
	s := len(data)
	elemType := field.Type().Elem()
	slice := reflect.MakeSlice(reflect.SliceOf(elemType), s, s)
	for i, v := range data {
		sv := slice.Index(i)
		kind := sv.Kind()
		if kind == reflect.Ptr {
			sv = reflect.New(elemType.Elem())
		} else {
			sv = sv.Addr()
		}
		tm := sv.Interface().(encoding.TextUnmarshaler)
		if err := tm.UnmarshalText([]byte(v)); err != nil {
			return fmt.Errorf(`env: parse error on field "%s" of type "%s": %v`, sf.Name, sf.Type, err)
		}
		if kind == reflect.Ptr {
			slice.Index(i).Set(sv)
		}
	}

	field.Set(slice)

	return nil
}

func getJsonType(item *simplejson.Json) string {
	switch item.Interface().(type) {
	case json.Number, int64, float64, int, int8, int16, int32, uint, uint8, uint16, uint32, uint64, float32:
		return "int"
	case bool:
		return "bool"
	case string:
		return "string"
	}

	return ""
}

func applyRemoteConfigForJson(config interface{}, remoteServer RemoteConfServer) error {
	jsonBytes, _ := json.Marshal(config)
	rspJsonBody, _ := simplejson.NewJson(jsonBytes)
	remoteKeys := remoteServer.GetAllKeys()

	for i := 0; i < len(remoteKeys); i++ {
		remoteKey := remoteKeys[i]
		typeStr := getJsonType(rspJsonBody.Get(remoteKey))
		value := remoteServer.Get(remoteKey)
		switch typeStr {
		case "int":
			num, _ := strconv.Atoi(value)
			rspJsonBody.Set(remoteKey, num)
		case "bool":
			if value == "true" {
				rspJsonBody.Set(remoteKey, true)
			} else {
				rspJsonBody.Set(remoteKey, false)
			}
		case "string":
			rspJsonBody.Set(remoteKey, value)
		default:
			continue
		}
	}

	jsonBytes, _ = rspJsonBody.MarshalJSON()
	json.Unmarshal(jsonBytes, &config)

	jsonBytes, _ = json.Marshal(config)
	return nil
}

func updateForJson(config interface{}, remoteKey, value string) error {
	jsonBytes, _ := json.Marshal(config)
	rspJsonBody, _ := simplejson.NewJson(jsonBytes)

	typeStr := getJsonType(rspJsonBody.Get(remoteKey))
	switch typeStr {
	case "int":
		num, _ := strconv.Atoi(value)
		rspJsonBody.Set(remoteKey, num)
	case "bool":
		if value == "true" {
			rspJsonBody.Set(remoteKey, true)
		} else {
			rspJsonBody.Set(remoteKey, false)
		}
	case "string":
		rspJsonBody.Set(remoteKey, value)
	}

	jsonBytes, _ = rspJsonBody.MarshalJSON()
	json.Unmarshal(jsonBytes, &config)
	return nil
}
