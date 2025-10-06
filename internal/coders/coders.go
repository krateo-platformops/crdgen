package coders

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/krateoplatformops/plumbing/env"
)

type Resource struct {
	Group        string
	Version      string
	Kind         string
	Categories   []string
	SpecSchema   []byte
	StatusSchema []byte
	Managed      bool
}

func ModuleName(kind string) string {
	return fmt.Sprintf("github.com/krateoplatformops/%s-crdgen", strings.ToLower(kind))
}

func GenAll(rootdir string, res *Resource) error {
	err := WriteTypesToFile(rootdir, res)
	if err != nil {
		return err
	}

	err = WriteGroupVersionInfoToFile(rootdir, res)
	return err
}

func WriteGroupVersionInfoToFile(rootdir string, opts *Resource) error {
	mod := ModuleName(opts.Kind)

	parts := []string{rootdir}
	parts = append(parts, strings.Split(mod, "/")...)
	parts = append(parts, "apis", opts.Version)

	workdir := filepath.Join(parts...)
	err := os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	bin, err := GenGroupVersionInfo(opts)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(workdir, "group_version_info.go"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(bin))
	return err
}

func WriteTypesToFile(rootdir string, opts *Resource) error {
	mod := ModuleName(opts.Kind)

	parts := []string{rootdir}
	parts = append(parts, strings.Split(mod, "/")...)
	parts = append(parts, "apis", opts.Version)

	workdir := filepath.Join(parts...)
	err := os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	bin, err := GenTypes(opts)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(workdir, "types.go"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(bin))
	return err
}

func GenTypes(opts *Resource) (dat []byte, err error) {
	co := newTypesCoder()

	err = co.parseSchemaForSpec(opts.SpecSchema)
	if err != nil {
		return nil, err
	}

	err = co.parseSchemaForStatus(opts.StatusSchema)
	if err != nil {
		return nil, err
	}

	co.addImports(opts.Version, opts.Managed)

	err = co.buildStructForDefs()
	if err != nil {
		return nil, err
	}

	err = co.buildStructForSpec(opts.Kind)
	if err != nil {
		return nil, err
	}

	err = co.buildStructForStatus(opts.Kind)
	if err != nil {
		return nil, err
	}

	co.buildEntryItemStructs(opts.Kind, opts.Categories, opts.Managed)
	co.buildEntryListStructs(opts.Kind, opts.Categories, opts.Managed)

	return co.bytes(env.True("FORMAT"))
}

func GenGroupVersionInfo(opts *Resource) (dat []byte, err error) {
	co := newGroupVersionInfoCoder()

	co.addImports(opts.Group, opts.Version)
	co.addConst(opts.Group, opts.Version)

	co.addVars(opts.Kind, opts.Group, opts.Version)

	co.initFunc(opts.Kind)

	return co.bytes(env.True("FORMAT"))
}
