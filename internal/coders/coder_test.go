package coders

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestCoder(t *testing.T) {
	ss, err := os.ReadFile("../../testdata/git.spec.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	res := Resource{
		Group:      "krateo.io",
		Version:    "v1alpha1",
		Kind:       "Provalo",
		Categories: []string{"krateo"},
		SpecSchema: ss,
		Managed:    true,
	}

	dat, err := Code(&res)
	if err != nil {
		t.Fatal(err)
	}

	io.Copy(os.Stdout, bytes.NewReader(dat))
}
