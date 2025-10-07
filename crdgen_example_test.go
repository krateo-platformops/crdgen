package crdgen_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/crdgen/v2"
	"github.com/krateoplatformops/crdgen/v2/internal/coders"
)

func TestGenerate(t *testing.T) {
	os.Setenv("KEEP_CODE", "1")

	specSchemaBytes, err := os.ReadFile("./testdata/git.spec.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	statusSchemaBytes, err := os.ReadFile("./testdata/git.status.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	yml, err := crdgen.Generate(coders.Options{
		Group:        "git.krateo.io",
		Version:      "v1alpha1",
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
