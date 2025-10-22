package crdgen

import (
	"os"

	"github.com/krateoplatformops/crdgen/v2/internal/coders"
	"github.com/krateoplatformops/crdgen/v2/internal/tools"
	"github.com/krateoplatformops/plumbing/env"
)

type Options struct {
	Group        string
	Version      string
	Kind         string
	Categories   []string
	SpecSchema   []byte
	StatusSchema []byte
	Managed      bool
}

func Generate(opts Options) (dat []byte, err error) {
	os.Setenv(coders.EnvFormatCode, "1")

	rootdir := os.TempDir()

	err = coders.GenAll(rootdir, &coders.Options{
		Group:        opts.Group,
		Version:      opts.Version,
		Kind:         opts.Kind,
		Categories:   opts.Categories,
		SpecSchema:   opts.SpecSchema,
		StatusSchema: opts.StatusSchema,
		Managed:      opts.Managed,
	})
	if err != nil {
		return
	}

	srcdir := coders.SourceDir(rootdir, opts.Kind)
	if !env.True(coders.EnvKeepCode) {
		defer os.RemoveAll(srcdir)
	}

	err = tools.Tidy(srcdir)
	if err != nil {
		return
	}

	return tools.GenerateCRDs(srcdir)
}
