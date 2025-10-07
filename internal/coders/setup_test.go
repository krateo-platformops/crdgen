package coders

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestGenSetup(t *testing.T) {
	os.Setenv("FORMAT", "1")

	res := Options{
		Group:   "git.krateo.io",
		Version: "v1alpha1",
		Kind:    "Repo",
	}

	dat, err := GenSetup(&res)
	if err != nil {
		t.Fatal(err)
	}

	io.Copy(os.Stdout, bytes.NewReader(dat))
}
