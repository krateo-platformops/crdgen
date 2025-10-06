package coders

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestGenTypes(t *testing.T) {
	os.Setenv("FORMAT", "1")

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

	dat, err := GenTypes(&res)
	if err != nil {
		t.Fatal(err)
	}

	io.Copy(os.Stdout, bytes.NewReader(dat))
}
