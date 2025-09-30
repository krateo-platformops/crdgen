package coders

import (
	"os"
	"testing"
)

func TestCodeForSpec(t *testing.T) {
	ss, err := os.ReadFile("../../testdata/array.enums.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	res := Resource{
		Group:      "templates.krateo.io",
		Version:    "v1",
		Kind:       "XApp",
		Categories: []string{"krateo", "rest"},
		SpecSchema: ss,
		Managed:    true,
	}

	dat, err := CodeForSpec(&res)
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout.Write(dat)
}
