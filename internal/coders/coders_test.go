package coders

import (
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/crdgen/v2/internal/tools"
)

func TestGenAll(t *testing.T) {
	os.Setenv("FORMAT", "1")

	rootdir := os.TempDir()

	specSchemaBytes, err := os.ReadFile("../../testdata/array.enums.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	var statusSchemaBytes []byte
	// statusSchemaBytes, err = os.ReadFile("../../testdata/git.status.schema.json")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	res := Options{
		Group:        "git.krateo.io",
		Version:      "v1alpha1",
		Kind:         "Repo",
		Categories:   []string{"krateo", "git", "repo"},
		SpecSchema:   specSchemaBytes,
		StatusSchema: statusSchemaBytes,
		Managed:      false,
	}

	err = GenAll(rootdir, &res)
	if err != nil {
		t.Fatal(err)
	}

	srcdir := SourceDir(rootdir, res.Kind)
	defer os.RemoveAll(srcdir)

	err = tools.Tidy(srcdir)
	if err != nil {
		t.Fatal(err)
	}
	yml, err := tools.GenerateCRDs(srcdir)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(yml))
}
