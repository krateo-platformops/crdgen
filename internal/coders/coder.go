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

type Resource struct {
	Group        string
	Version      string
	Kind         string
	Categories   []string
	SpecSchema   []byte
	StatusSchema []byte
	Managed      bool
}

func Code(opts *Resource) (dat []byte, err error) {
	co := newCoder()

	err = co.parseSchemaForSpec(opts.SpecSchema)
	if err != nil {
		return nil, err
	}

	err = co.parseSchemaForStatus(opts.StatusSchema)
	if err != nil {
		return nil, err
	}

	err = co.buildStructForDefs()
	if err != nil {
		return nil, err
	}

	err = co.buildStructForSpec(opts.Kind)
	if err != nil {
		return nil, err
	}

	err = co.buildStructForStatus(opts.Kind)
	if err != nil {
		return nil, err
	}

	return co.Bytes()
}

/*
	func Code2(gen *gg.Generator, opts *Resource) (dat []byte, err error) {
		var (
			specSchema   *schemas.Schema
			statusSchema *schemas.Schema
			defs         map[string]*schemas.Type
		)

		defs = map[string]*schemas.Type{}

		if opts.SpecSchema != nil {
			specSchema, err = schemas.FromJSONReader(bytes.NewReader(opts.SpecSchema))
			if err != nil {
				return nil, err
			}
			specDefs := schemas.CollectAllDefinitions(specSchema)
			maps.Copy(defs, specDefs)
		}

		if opts.StatusSchema != nil {
			statusSchema, err = schemas.FromJSONReader(bytes.NewReader(opts.StatusSchema))
			if err != nil {
				return nil, err
			}
			statusDefs := schemas.CollectAllDefinitions(statusSchema)
			maps.Copy(defs, statusDefs)
		}

		// Risolvi allOf per tutte le definitions
		resolvedDefs := make(map[string]*schemas.Type)

		for name, def := range defs {
			resolved := def
			if len(def.AllOf) > 0 {
				merged, err := schemas.AllOf(def.AllOf, specSchema.Definitions)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve allOf for %s: %w", name, err)
				}
				resolved = merged
			}
			//resolved = resolveRefsInType(resolved, defs)
			resolvedDefs[name] = resolved

			buildStruct(gen, name, resolved, resolvedDefs)
		}

		// 2️⃣ Gen code for the root type (non-definitions)
		rootName := opts.Kind
		if rootName == "" {
			rootName = "Spec" // default
		}

		rootType := schemaAsType(specSchema)
		//rootType = resolveRefsInType(rootType, resolvedDefs)
		if len(rootType.Properties) > 0 {
			if err := buildStruct(gen, opts.Kind, rootType, resolvedDefs); err != nil {
				return nil, err
			}
		}

		buf := bytes.Buffer{}
		gen.Write(&buf)

		return format.Source(buf.Bytes())
	}
*/
func newCoder() *coder {
	return &coder{
		gen:              gg.New(),
		resolvedDefs:     map[string]*schemas.Type{},
		generatedStructs: map[string]bool{},
		generatedEnums:   map[string]bool{},
	}
}

type coder struct {
	gen              *gg.Generator
	specSchema       *schemas.Schema
	statusSchema     *schemas.Schema
	resolvedDefs     map[string]*schemas.Type
	generatedStructs map[string]bool
	generatedEnums   map[string]bool
}

func (co *coder) Bytes() ([]byte, error) {
	buf := bytes.Buffer{}
	co.gen.Write(&buf)

	return format.Source(buf.Bytes())
}

func (co *coder) parseSchemaForSpec(in []byte) (err error) {
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

func (co *coder) parseSchemaForStatus(in []byte) (err error) {
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

func (co *coder) buildStructForDefs() (err error) {
	for name, def := range co.resolvedDefs {
		err = co.buildStruct(name, def)
		if err != nil {
			return
		}
	}

	return nil
}

func (co *coder) resolveAllOf(in *schemas.Schema, defs map[string]*schemas.Type) error {
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

func (co *coder) buildStructForSpec(kind string) (err error) {
	if co.specSchema == nil {
		return
	}

	rootName := kind
	if rootName == "" {
		rootName = "Spec" // default
	}

	rootType := schemaAsType(co.specSchema)
	//rootType = resolveRefsInType(rootType, resolvedDefs)
	if len(rootType.Properties) > 0 {
		err = co.buildStruct(rootName, rootType)
		if err != nil {
			return err
		}
	}

	return nil
}

func (co *coder) buildStructForStatus(kind string) (err error) {
	if co.statusSchema == nil {
		return
	}

	rootName := kind
	if rootName == "" {
		rootName = "Status" // default
	}

	rootType := schemaAsType(co.statusSchema)
	//rootType = resolveRefsInType(rootType, resolvedDefs)
	if len(rootType.Properties) > 0 {
		err = co.buildStruct(rootName, rootType)
		if err != nil {
			return err
		}
	}

	return nil
}

func (co *coder) buildStruct(typeName string, t *schemas.Type) error {
	if co.generatedStructs[typeName] {
		return nil // già generata
	}
	co.generatedStructs[typeName] = true

	st := co.gen.NewGroup().NewStruct(typeName)

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

		st.AddField(fieldName, fieldType, tags)
	}

	return nil
}

// helper per convertire $ref in nome struct
func refToTypeName(ref string) string {
	parts := strings.Split(ref, "/")
	return exportedName(parts[len(parts)-1])
}

func (co *coder) resolveType(typeName string, t *schemas.Type) string {
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
		co.buildStruct(refName, resolved)
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
			co.buildStruct(typeName, t)
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

func (co *coder) emitEnum(typeName string, t *schemas.Type) string {
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
