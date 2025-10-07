package crdgen

import (
	"os"

	"github.com/krateoplatformops/crdgen/v2/internal/coders"
	"github.com/krateoplatformops/crdgen/v2/internal/tools"
	"github.com/krateoplatformops/plumbing/env"
)

func Generate(opts coders.Options) (dat []byte, err error) {
	os.Setenv("FORMAT", "1")

	rootdir := os.TempDir()

	err = coders.GenAll(rootdir, &opts)
	if err != nil {
		return
	}

	srcdir := coders.SourceDir(rootdir, opts.Kind)
	if !env.True("KEEP_CODE") {
		defer os.RemoveAll(srcdir)
	}

	err = tools.Tidy(srcdir)
	if err != nil {
		return
	}

	return tools.GenerateCRDs(srcdir)
}
