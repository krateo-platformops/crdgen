package transpiler

import (
	"fmt"
	"reflect"
)

func strval(v any) string {
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

func strslice(v any) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []any:
		b := make([]string, 0, len(v))
		for _, s := range v {
			if s != nil {
				b = append(b, strval(s))
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
					b = append(b, strval(value))
				}
			}
			return b
		default:
			if v == nil {
				return []string{}
			}

			return []string{strval(v)}
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// reservedTypeNames are Go identifiers that crdgen emits as package-level
// declarations (constants, vars, funcs) in the generated apis package,
// alongside the struct types produced from the JSON schema. A generated
// struct type must never reuse one of these names or it would redeclare /
// shadow the package-level identifier and break code generation.
//
// The canonical failure is a schema whose top-level property is named
// "group" (e.g. the OpenStack Keystone group envelope): it yields
// `type Group struct{...}` which collides with `const Group = "<api group>"`
// emitted in groupversion_info.go.
var reservedTypeNames = map[string]struct{}{
	"Group":              {},
	"Version":            {},
	"SchemeGroupVersion": {},
	"SchemeBuilder":      {},
	"AddToScheme":        {},
	"AddToSchemes":       {},
}

// safeTypeName returns a Go type name that does not collide with the
// package-level identifiers emitted by crdgen. Colliding names are given an
// "Envelope" suffix (e.g. "Group" -> "GroupEnvelope"), which is stable and
// idempotent and avoids clashing with the <Kind>Spec/<Kind>Status/<Kind>List
// types that crdgen also generates.
func safeTypeName(name string) string {
	if _, reserved := reservedTypeNames[name]; reserved {
		return name + "Envelope"
	}
	return name
}
