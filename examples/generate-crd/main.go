// Command generate-crd shows the ONE way this library is consumed: feed a chart's
// values.schema.json (the spec) to crdgen.Generate and get back a composition CRD
// manifest. Run it with `go run .` from this directory — it prints the CRD YAML on
// stdout. Requires a Go toolchain on PATH and network access: Generate materializes
// a throwaway Go module in a temp dir and shells out to `go mod tidy` + controller-gen.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/krateo-platformops/crdgen"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// fileJsonSchemaGetter satisfies crdgen.JsonSchemaGetter by reading a JSON Schema
// document from disk — the same shape a consumer uses to feed a chart's
// values.schema.json pulled from any source (file, OCI layer, HTTP).
type fileJsonSchemaGetter struct {
	filename string
}

func (g *fileJsonSchemaGetter) Get() ([]byte, error) {
	fin, err := os.Open(g.filename)
	if err != nil {
		return nil, err
	}
	defer fin.Close()
	return io.ReadAll(fin)
}

func main() {
	res := crdgen.Generate(context.Background(), crdgen.Options{
		// WorkDir names the throwaway module dir under os.TempDir().
		WorkDir: "sampleapp",
		// The chart identity: Kind is the chart name in PascalCase, Version the
		// chart version in CRD-legal form (dots -> dashes, e.g. 1.2.3 -> v1-2-3).
		GVK: schema.GroupVersionKind{
			Group:   "composition.krateo.io",
			Version: "v1-2-3",
			Kind:    "SampleApp",
		},
		Categories: []string{"compositions", "comps"},
		// Managed adds the provider-runtime ConditionedStatus (Ready condition,
		// failedObjectRef) to .status and the READY printer column.
		Managed:              true,
		SpecJsonSchemaGetter: &fileJsonSchemaGetter{"./values.schema.json"},
	})
	if res.Err != nil {
		fmt.Fprintln(os.Stderr, "generation failed:", res.Err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "generated %s (spec digest %s)\n", res.GVK, res.Digest)
	fmt.Println(string(res.Manifest))
}
