package coders

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"

	gg "github.com/krateoplatformops/crdgen/v2/internal/codegen"
	"github.com/krateoplatformops/crdgen/v2/internal/schemas"
	ptrutils "github.com/krateoplatformops/crdgen/v2/internal/utils/ptr"
	stringsutils "github.com/krateoplatformops/crdgen/v2/internal/utils/strings"
)

func CodeForSpec(res *Resource) (code []byte, err error) {
	sch, err := schemas.FromJSONReader(bytes.NewReader(res.SpecSchema))
	if err != nil {
		return nil, err
	}

	g := gg.New()

	g.NewGroup().AddPackage(res.Version).NewImport().
		AddAlias("k8s.io/apimachinery/pkg/apis/meta/v1", "metav1").
		AddPath("k8s.io/apimachinery/pkg/runtime")

	buildSpecStruct(g, res.Kind+"Spec", (*schemas.Type)(sch.ObjectAsType))

	tmp := g.NewGroup()
	tmp.AddLineComment("+kubebuilder:object:root=true")

	tmp.AddLineComment("+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object")
	tmp.AddLineComment("+kubebuilder:subresource:status")
	tmp.AddLineComment("+kubebuilder:printcolumn:name=\"AGE\",type=\"date\",JSONPath=\".metadata.creationTimestamp\"")
	if len(res.Categories) > 0 {
		categories := strings.Join(res.Categories, ",")
		tmp.AddLineComment(fmt.Sprintf(
			"+kubebuilder:resource:scope=Namespaced,categories={%s}", categories))
	} else {
		tmp.AddLineComment("+kubebuilder:resource:scope=Namespaced")
	}

	st := tmp.NewStruct(res.Kind)
	st.AddField("", "metav1.TypeMeta", map[string]string{
		"json": ",inline",
	})
	st.AddField("", "metav1.ObjectMeta", map[string]string{
		"json": "metadata",
	})
	st.AddField("Spec", res.Kind+"Spec", map[string]string{
		"json": "spec",
	})
	if res.Managed {
		if res.StatusSchema == nil {
			st.AddField("Status", "*runtime.RawExtension", map[string]string{
				"json": "status,omitempty",
			})
		}
	}

	buf := bytes.Buffer{}
	g.Write(&buf)

	code, err = format.Source(buf.Bytes())
	return
}

func buildSpecStruct(g *gg.Generator, typeName string, t *schemas.Type) {
	code := g.NewGroup()

	st := code.NewStruct(typeName)

	for name, prop := range t.Properties {
		fieldName := exportedName(name)
		fieldType := resolveType(g, fieldName, prop)

		if isRequired(t, fieldName) {
			st.AddLineComment("+required")
		}

		if prop.Title != "" {
			st.AddLineComment("+kubebuilder:title:%s", prop.Title)
		}

		if prop.Default != nil {
			st.AddLineComment("+kubebuilder:default:%s", stringsutils.StrVal(prop.Default))
		}

		if prop.Minimum != nil {
			st.AddLineComment("+kubebuilder:validation:Minimum:%s",
				stringsutils.StrVal(ptrutils.Deref(prop.Minimum, 0)))
		}

		if prop.Maximum != nil {
			st.AddLineComment("+kubebuilder:validation:Maximum:%s",
				stringsutils.StrVal(ptrutils.Deref(prop.Maximum, 0)))
		}

		if prop.MultipleOf != nil {
			st.AddLineComment("+kubebuilder:validation:MultipleOf:%s",
				stringsutils.StrVal(ptrutils.Deref(prop.MultipleOf, 0)))
		}

		if prop.Pattern != "" {
			st.AddLineComment("+kubebuilder:validation:Pattern:%s", prop.Pattern)
		}

		st.AddField(fieldName, fieldType,
			map[string]string{
				"json": fmt.Sprintf("%s,omitempty", name),
			})
	}
}

func resolveType(g *gg.Generator, typeName string, t *schemas.Type) string {
	// Caso 1: enum string semplice
	if t.Type.Equals(schemas.TypeList{"string"}) && len(t.Enum) > 0 {
		return emitEnum(g, typeName, t)
	}

	// Caso 2: array
	if t.Type.Equals(schemas.TypeList{"array"}) {
		if t.Items != nil {
			// se gli items sono stringhe con enum → alias + costanti
			if t.Items.Type.Equals(schemas.TypeList{"string"}) && len(t.Items.Enum) > 0 {
				itemType := emitEnum(g, typeName+"Item", t.Items)
				return "[]" + itemType
			}
			// altrimenti normale array
			itemType := resolveType(g, typeName+"Item", t.Items)
			return "[]" + itemType
		}
		return "[]runtime.RawExtension"
	}

	// Caso 3: oggetti con proprietà
	if t.Type.Equals(schemas.TypeList{"object"}) && len(t.Properties) > 0 {
		buildSpecStruct(g, typeName, t)
		return typeName
	}

	// Caso 4: fallback tabellare
	return jsonSchemaToGoType(t)
}

func emitEnum(g *gg.Generator, typeName string, t *schemas.Type) string {
	grp := g.NewGroup()
	grp.AddLineComment("+kubebuilder:validation:Enum:=" + stringsutils.Join(t.Enum, ";"))
	grp.AddTypeAlias(typeName, "string")

	consts := g.NewGroup()
	for _, e := range t.Enum {
		if s, ok := e.(string); ok {
			constName := typeName + exportedName(s)
			consts.NewConst().AddTypedField(constName, typeName, gg.Lit(s))
		}
	}
	return typeName
}

// exportedName trasforma un nome JSON in un identificatore Go esportato
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
