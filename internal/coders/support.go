package coders

import (
	"strings"

	"github.com/krateoplatformops/crdgen/v2/internal/schemas"
)

func schemaAsType(s *schemas.Schema) *schemas.Type {
	return &schemas.Type{
		Type:                 s.Type,
		Properties:           s.Properties,
		Items:                s.Items,
		Enum:                 s.Enum,
		Format:               s.Format,
		Required:             s.Required,
		AdditionalProperties: s.AdditionalProperties,
		AllOf:                s.AllOf,
		OneOf:                s.OneOf,
		AnyOf:                s.AnyOf,
		Default:              s.Default,
		Minimum:              s.Minimum,
		Maximum:              s.Maximum,
		MultipleOf:           s.MultipleOf,
		Pattern:              s.Pattern,
		Definitions:          s.Definitions,
		Ref:                  s.Ref,
	}
}

func resolveRefDefs(t *schemas.Type, defs schemas.Definitions, visited map[string]bool) (*schemas.Type, error) {
	if t == nil {
		return nil, nil
	}

	if t.Ref != "" {
		refName := strings.TrimPrefix(t.Ref, "#/$defs/")
		if visited[refName] {
			return t, nil
		}
		visited[refName] = true

		if resolved, ok := defs[refName]; ok {
			return resolveRefDefs(resolved, defs, visited)
		}
		// fallback
		return t, nil
	}

	if len(t.AllOf) > 0 {
		resolvedTypes := make([]*schemas.Type, len(t.AllOf))
		for i, sub := range t.AllOf {
			r, err := resolveRefDefs(sub, defs, visited)
			if err != nil {
				return nil, err
			}
			resolvedTypes[i] = r
		}
		return schemas.AllOf(resolvedTypes, defs)
	}

	if t.Properties != nil {
		newProps := make(map[string]*schemas.Type, len(t.Properties))
		for k, v := range t.Properties {
			r, err := resolveRefDefs(v, defs, visited)
			if err != nil {
				return nil, err
			}
			newProps[k] = r
		}
		t.Properties = newProps
	}

	return t, nil
}

func isNullable(t *schemas.Type) bool {
	for _, typ := range t.Type {
		if typ == "null" {
			return true
		}
	}
	return false
}

func isRequired(schema *schemas.Type, key string) bool {
	if schema == nil {
		return false
	}

	for _, el := range schema.Required {
		if strings.EqualFold(key, el) {
			return true
		}
	}

	return false
}

// jsonSchemaToGoType converte un JSON Schema type/format in un tipo Go compatibile CRD
func jsonSchemaToGoType(t *schemas.Type) string {
	switch {
	case t.Type.Equals(schemas.TypeList{"string"}):
		switch t.Format {
		case "date-time":
			return "metav1.Time"
		case "duration":
			return "metav1.Duration"
		case "quantity":
			return "resource.Quantity"
		default:
			return "string"
		}

	case t.Type.Equals(schemas.TypeList{"boolean"}):
		return "bool"

	case t.Type.Equals(schemas.TypeList{"integer"}):
		// se lo schema specifica format=int64 → usa int64
		if t.Format == "int64" {
			return "int64"
		}
		return "int32"

	case t.Type.Equals(schemas.TypeList{"number"}):
		return "float64"

	case t.Type.Equals(schemas.TypeList{"array"}):
		if t.Items != nil {
			itemType := jsonSchemaToGoType(t.Items)
			return "[]" + itemType
		}
		return "[]runtime.RawExtension"

	case t.Type.Equals(schemas.TypeList{"object"}):
		if t.AdditionalProperties != nil {
			valType := jsonSchemaToGoType(t.AdditionalProperties)
			return "map[string]" + valType
		}
		if len(t.Properties) > 0 {
			// struct → deve essere costruita altrove (es. buildStruct)
			// qui ritorniamo un placeholder
			return ""
		}
		return "runtime.RawExtension"
	}

	// fallback → JSON arbitrario
	return "runtime.RawExtension"
}

func exportedName(name string) string {
	if name == "" {
		return name
	}

	parts := strings.Split(name, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}
