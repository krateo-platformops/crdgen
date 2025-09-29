package schemas

import (
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestParse(t *testing.T) {
	const (
		js = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "type": "object",
    "properties": {
        "maybeAllowedResources": {
            "description": "the list of resources that are allowed to be children of this widget or referenced by it",
            "type": "array",
            "items": {
                "type": "string",
                "enum": [
                    "blue", "green", "red", "black"
                ]
            },
             "default": ["blue", "green", "red"]
        }
    }
}`
	)

	sch, err := FromJSONReader(strings.NewReader(js))
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(sch)
}
