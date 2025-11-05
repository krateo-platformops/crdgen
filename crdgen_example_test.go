//go:build integration
// +build integration

package crdgen_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/crdgen/v2"
)

func TestGenerate(t *testing.T) {
	//os.Setenv("KEEP_CODE", "1")

	specSchemaBytes, err := os.ReadFile("./testdata/array.enums.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	statusSchemaBytes, err := os.ReadFile("./testdata/git.status.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	yml, err := crdgen.Generate(crdgen.Options{
		Group:        "git.krateo.io",
		Version:      "v1.2.3",
		Kind:         "Repo",
		Categories:   []string{"krateo", "git", "repo"},
		SpecSchema:   specSchemaBytes,
		StatusSchema: statusSchemaBytes,
		Managed:      true,
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(yml))
}

func TestGenerateButtonSchema(t *testing.T) {
	os.Setenv("KEEP_CODE", "1")

	const (
		widgetsGroup          = "widgets.templates.krateo.io"
		preserveUnknownFields = `{"type": "object", "additionalProperties": true,"x-kubernetes-preserve-unknown-fields": true}`
	)

	specSchemaBytes, err := os.ReadFile("./testdata/button.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	yml, err := crdgen.Generate(crdgen.Options{
		Group:        widgetsGroup,
		Version:      "v0.0.7",
		Kind:         "Button",
		Categories:   []string{"krateo", "widgets"},
		SpecSchema:   specSchemaBytes,
		StatusSchema: []byte(preserveUnknownFields),
	})
	if err != nil {
		t.Fatal(err)
	}

	// f, _ := os.Create("vcluster.yaml")
	// defer f.Close()

	//io.Copy(f, strings.NewReader(string(yml)))
	fmt.Println(string(yml))
}

func TestGenerateGithubScaffolding(t *testing.T) {
	//os.Setenv("KEEP_CODE", "1")

	const (
		widgetsGroup = "github.krateo.io"
	)

	specSchemaBytes, err := os.ReadFile("./testdata/github-scaffolding.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	yml, err := crdgen.Generate(crdgen.Options{
		Group:      widgetsGroup,
		Version:    "v0.0.7",
		Kind:       "Scaffolding",
		Categories: []string{"krateo", "github"},
		SpecSchema: specSchemaBytes,
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(yml))
}

func TestIssueWrongCaseFields(t *testing.T) {
	//os.Setenv("KEEP_CODE", "1")

	const (
		js = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "type": "object",
    "properties": {
      "greet-ing": {
        "type": "string"
      },
      "dis_play": {
        "type": "integer",
        "default": 1
      },
      "verb--ose": {
        "type": "boolean",
        "default": false
      },
      "url": {
        "type": "string",
        "default": "https://github.com/krateoplatformops/sticz"
      }
    },
    "required": [
      "greeting"
    ]
  }`
	)

	yml, err := crdgen.Generate(crdgen.Options{
		Group:      "krateo.io",
		Version:    "v1alpha1",
		Kind:       "Hello",
		Categories: []string{"krateo", "hello"},
		SpecSchema: []byte(js),
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(yml))
}
