package coders

import (
	"fmt"
	"os"
	"testing"
)

func TestGenAll(t *testing.T) {
	os.Setenv("FORMAT", "1")

	rootdir := os.TempDir()

	specSchemaBytes, err := os.ReadFile("../../testdata/git.spec.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	statusSchemaBytes, err := os.ReadFile("../../testdata/git.status.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	res := Resource{
		Group:        "git.krateo.io",
		Version:      "v1alpha1",
		Kind:         "Repo",
		Categories:   []string{"krateo", "git", "repo"},
		SpecSchema:   specSchemaBytes,
		StatusSchema: statusSchemaBytes,
		Managed:      true,
	}

	err = GenAll(rootdir, &res)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(rootdir)
}
