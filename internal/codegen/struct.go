package codegen

import (
	"io"
)

type istruct struct {
	name  string
	items *group
}

// Struct will insert a new struct.
func Struct(name string) *istruct {
	return &istruct{
		name: name,
		// We will insert new line before closing the struct to avoid being affect
		// by line comments.
		items: newGroup("{\n", "\n}", "\n"),
	}
}

func (i *istruct) render(w io.Writer) {
	writeStringF(w, "type %s struct ", i.name)
	i.items.render(w)
}

// AddLine will insert an empty line.
func (i *istruct) AddLine() *istruct {
	i.items.append(Line())
	return i
}

// AddLineComment will insert a new line comment.
func (i *istruct) AddLineComment(content string, args ...any) *istruct {
	i.items.append(LineComment(content, args...))
	return i
}

func (i *istruct) AddField(name, typ any, tags map[string]string) *istruct {
	i.items.append(field(name, typ, " ", tags))
	return i
}
