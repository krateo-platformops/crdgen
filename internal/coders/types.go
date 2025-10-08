package coders

import (
	"bytes"
	"fmt"
	"go/format"
	"slices"
	"strings"

	gg "github.com/krateoplatformops/crdgen/v2/internal/codegen"
	"github.com/krateoplatformops/crdgen/v2/internal/schemas"
	ptrutils "github.com/krateoplatformops/crdgen/v2/internal/utils/ptr"
	stringsutils "github.com/krateoplatformops/crdgen/v2/internal/utils/strings"
)

func newTypesCoder() *typesCoder {
	return &typesCoder{
		gen:              gg.New(),
		resolvedDefs:     map[string]*schemas.Type{},
		generatedStructs: map[string]bool{},
		generatedEnums:   map[string]bool{},
	}
}

type typesCoder struct {
	gen              *gg.Generator
	specSchema       *schemas.Schema
	statusSchema     *schemas.Schema
	resolvedDefs     map[string]*schemas.Type
	generatedStructs map[string]bool
	generatedEnums   map[string]bool
}

func (co *typesCoder) bytes(gofmt bool) ([]byte, error) {
	buf := bytes.Buffer{}
	co.gen.Write(&buf)

	if gofmt {
		return format.Source(buf.Bytes())
	}

	return buf.Bytes(), nil
}

func (co *typesCoder) parseSchemaForSpec(in []byte) (err error) {
	if in == nil {
		return
	}

	co.specSchema, err = schemas.FromJSONReader(bytes.NewReader(in))
	if err != nil {
		return err
	}

	if co.specSchema == nil {
		return
	}

	defs := schemas.CollectAllDefinitions(co.specSchema)

	return co.resolveAllOf(co.specSchema, defs)
}

func (co *typesCoder) parseSchemaForStatus(in []byte) (err error) {
	if in == nil {
		return
	}

	co.statusSchema, err = schemas.FromJSONReader(bytes.NewReader(in))
	if err != nil {
		return
	}

	if co.statusSchema == nil {
		return
	}

	defs := schemas.CollectAllDefinitions(co.statusSchema)

	err = co.resolveAllOf(co.statusSchema, defs)
	return err
}

func (co *typesCoder) buildStructForDefs() (err error) {
	for name, def := range co.resolvedDefs {
		err = co.buildStruct(name, def, nil)
		if err != nil {
			return
		}
	}

	return nil
}

func (co *typesCoder) resolveAllOf(in *schemas.Schema, defs map[string]*schemas.Type) error {
	for name, def := range defs {
		resolved := def
		if len(def.AllOf) > 0 {
			merged, err := schemas.AllOf(def.AllOf, in.Definitions)
			if err != nil {
				return fmt.Errorf("failed to resolve allOf for %s: %w", name, err)
			}
			resolved = merged
		}

		co.resolvedDefs[name] = resolved
	}

	return nil
}

func (co *typesCoder) buildStructForSpec(kind string) (err error) {
	if co.specSchema == nil {
		return
	}

	rootName := kind + "Spec"
	rootType := schemaAsType(co.specSchema)
	//rootType = resolveRefsInType(rootType, resolvedDefs)
	if len(rootType.Properties) > 0 {
		err = co.buildStruct(rootName, rootType, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (co *typesCoder) buildStructForStatus(kind string) (err error) {
	if co.statusSchema == nil {
		return
	}

	rootName := kind + "Status"
	rootType := schemaAsType(co.statusSchema)
	//rootType = resolveRefsInType(rootType, resolvedDefs)
	if len(rootType.Properties) > 0 {
		err = co.buildStruct(rootName, rootType, func(st *gg.IStruct) {
			st.AddField("commonv1.ConditionedStatus", "",
				map[string]string{"json": ",inline"})
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (co *typesCoder) addImports(version string, managed bool) {
	pkgs := co.gen.NewGroup().AddPackage(version).NewImport().
		AddAlias("k8s.io/apimachinery/pkg/apis/meta/v1", "metav1").
		AddPath("k8s.io/apimachinery/pkg/runtime")

	if managed {
		pkgs.AddAlias("github.com/krateoplatformops/provider-runtime/apis/common/v1", "commonv1")
		pkgs.AddPath("github.com/krateoplatformops/provider-runtime/pkg/resource")
	}
}

func (co *typesCoder) buildEntryItemStructs(kind string, categories []string, managed bool) {
	grp := co.gen.NewGroup().AddLine()
	grp.AddLineComment("+kubebuilder:object:root=true")

	grp.AddLineComment("+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object")
	if co.statusSchema != nil {
		grp.AddLineComment("+kubebuilder:subresource:status")
	}

	grp.AddLineComment(
		`+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"`)
	if managed {
		grp.AddLineComment(
			`+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"`)
	}

	if len(categories) > 0 {
		grp.AddLineComment(fmt.Sprintf(
			"+kubebuilder:resource:scope=Namespaced,categories={%s}", strings.Join(categories, ",")))
	} else {
		grp.AddLineComment("+kubebuilder:resource:scope=Namespaced")
	}

	st := co.gen.NewGroup().NewStruct(kind)
	st.AddField("", "metav1.TypeMeta", map[string]string{
		"json": ",inline",
	})
	st.AddField("", "metav1.ObjectMeta", map[string]string{
		"json": "metadata",
	})
	st.AddField("Spec", kind+"Spec", map[string]string{
		"json": "spec",
	})

	if co.statusSchema != nil {
		st.AddField("Status", kind+"Status", map[string]string{
			"json": "status",
		})
	}

	if !managed {
		return
	}

	grp = co.gen.NewGroup().AddLineComment("GetCondition of this %s", kind)
	grp.NewFunction("GetCondition").
		WithReceiver("mg", "*"+kind).
		AddParameter("ct", "commonv1.ConditionType").
		AddResult("", "commonv1.Condition").
		AddBody(gg.String("return mg.Status.GetCondition(ct)"))

	grp = co.gen.NewGroup().AddLineComment("SetConditions of this %s", kind)
	grp.NewFunction("SetConditions").
		WithReceiver("mg", "*"+kind).
		AddParameter("c", "...commonv1.Condition").
		AddBody(gg.String("mg.Status.SetConditions(c...)"))

}

func (co *typesCoder) buildEntryListStructs(kind string, managed bool) {
	name := kind + "List"

	grp := co.gen.NewGroup().AddLine()
	grp.AddLineComment("+kubebuilder:object:root=true")

	st := co.gen.NewGroup().NewStruct(name)
	st.AddField("", "metav1.TypeMeta", map[string]string{
		"json": ",inline",
	})
	st.AddField("", "metav1.ListMeta", map[string]string{
		"json": "metadata,omitempty",
	})
	st.AddField("Items", "[]"+kind, map[string]string{
		"json": "items",
	})

	if !managed {
		return
	}

	grp = co.gen.NewGroup().AddLineComment("GetItems of this %s", name)
	grp.NewFunction("GetItems").
		AddResult("", "[]resource.Managed").
		WithReceiver("l", "*"+name).
		AddBody(gg.String("items := make([]resource.Managed, len(l.Items))")).
		AddBody("for i := range l.Items {").
		AddBody("items[i] = &l.Items[i]").
		AddBody("}").
		AddBody("return items")
}

func (co *typesCoder) buildStruct(typeName string, t *schemas.Type, cb func(*gg.IStruct)) error {
	if co.generatedStructs[typeName] {
		return nil // già generata
	}
	co.generatedStructs[typeName] = true

	st := co.gen.NewGroup().NewStruct(typeName)
	if cb != nil {
		cb(st)
	}

	for name, prop := range t.Properties {
		fieldName := exportedName(name)

		fieldType := co.resolveType(fieldName, prop)

		if isNullable(prop) && !strings.HasPrefix(fieldType, "*") && fieldType != "runtime.RawExtension" {
			fieldType = "*" + fieldType
		}

		// tag json
		tags := map[string]string{}
		if isNullable(prop) || strings.HasPrefix(fieldType, "*") {
			tags["json"] = fmt.Sprintf("%s,omitempty", name)
		} else {
			tags["json"] = name
		}

		// kubebuilder annotations
		if isRequired(t, name) {
			st.AddLineComment("+required")
		}
		if prop.Title != "" {
			st.AddLineComment("+kubebuilder:title:%s", prop.Title)
		}
		if prop.Default != nil {
			st.AddLineComment("+kubebuilder:default:=%s", stringsutils.DefaultValForKubebuilder(prop.Default))
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

		if prop.Description != "" {
			st.AddLineComment(prop.Description)
		}

		st.AddField(fieldName, fieldType, tags)
	}

	return nil
}

// helper per convertire $ref in nome struct
func refToTypeName(ref string) string {
	parts := strings.Split(ref, "/")
	return exportedName(parts[len(parts)-1])
}

func (co *typesCoder) resolveType(typeName string, t *schemas.Type) string {
	// Caso $ref
	if t.Ref != "" {
		refName := refToTypeName(t.Ref)
		if co.generatedStructs[refName] {
			if !slices.Contains(t.Required, refName) {
				return "*" + refName
			}
			return refName
		}
		resolved, err := resolveRefDefs(t, co.resolvedDefs, map[string]bool{})
		if err != nil {
			return "runtime.RawExtension"
		}
		co.buildStruct(refName, resolved, nil)
		return refName
	}

	// Nullable: ["null", "type"]
	if isNullable(t) && len(t.Type) == 2 {
		nonNullType := &schemas.Type{Type: schemas.TypeList{}}
		for _, typ := range t.Type {
			if typ != "null" {
				nonNullType.Type = schemas.TypeList{typ}
				nonNullType.Properties = t.Properties
				nonNullType.Items = t.Items
				nonNullType.Enum = t.Enum
				nonNullType.Format = t.Format
				nonNullType.AdditionalProperties = t.AdditionalProperties
			}
		}
		base := co.resolveType(typeName, nonNullType)
		// Solo se non è già pointer
		if !strings.HasPrefix(base, "*") {
			base = "*" + base
		}
		return base
	}

	// enum
	if t.Type.Equals(schemas.TypeList{"string"}) && len(t.Enum) > 0 {
		return co.emitEnum(typeName, t)
	}

	// array
	if t.Type.Equals(schemas.TypeList{"array"}) {
		if t.Items != nil {
			itemType := co.resolveType(typeName+"Item", t.Items)
			return "[]" + itemType
		}
		return "[]runtime.RawExtension"
	}

	// oggetto con proprietà → genera struct
	if t.Type.Equals(schemas.TypeList{"object"}) && len(t.Properties) > 0 {
		if !co.generatedStructs[typeName] {
			co.buildStruct(typeName, t, nil)
		}
		return typeName
	}

	// Oggetto con AdditionalProperties (map)
	if t.Type.Equals(schemas.TypeList{"object"}) && t.AdditionalProperties != nil {
		valType := co.resolveType(typeName+"Value", t.AdditionalProperties)
		return "map[string]" + valType
	}

	// object con AdditionalProperties o fallback
	return jsonSchemaToGoType(t)
}

func (co *typesCoder) emitEnum(typeName string, t *schemas.Type) string {
	if co.generatedEnums[typeName] {
		return typeName
	}
	co.generatedEnums[typeName] = true

	grp := co.gen.NewGroup()
	grp.AddLineComment("+kubebuilder:validation:Enum:=" + stringsutils.Join(t.Enum, ";"))
	grp.AddTypeAlias(typeName, "string")

	consts := co.gen.NewGroup()
	for _, e := range t.Enum {
		if s, ok := e.(string); ok {
			constName := typeName + exportedName(s)
			consts.NewConst().AddTypedField(constName, typeName, gg.Lit(s))
		}
	}
	return typeName
}
