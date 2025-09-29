package strings

import (
	"fmt"
	"reflect"
	"strings"
)

func StrSlice(v any) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []any:
		b := make([]string, 0, len(v))
		for _, s := range v {
			if s != nil {
				b = append(b, StrVal(s))
			}
		}
		return b
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			l := val.Len()
			b := make([]string, 0, l)
			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				if value != nil {
					b = append(b, StrVal(value))
				}
			}
			return b
		default:
			if v == nil {
				return []string{}
			}

			return []string{StrVal(v)}
		}
	}
}

func StrVal(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func Join(v any, sep string) string {
	return strings.Join(StrSlice(v), sep)
}
