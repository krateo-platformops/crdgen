package crdgen

import (
	"io"
	"os"
)

type JsonSchemaGetter interface {
	Get() ([]byte, error)
}

var _ JsonSchemaGetter = (*fileJsonSchemaGetter)(nil)

type fileJsonSchemaGetter struct {
	filename string
}

func (f *fileJsonSchemaGetter) Get() ([]byte, error) {
	fin, err := os.Open(f.filename)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	return io.ReadAll(fin)
}
