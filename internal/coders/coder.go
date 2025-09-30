package coders

type Resource struct {
	Group        string
	Version      string
	Kind         string
	Categories   []string
	SpecSchema   []byte
	StatusSchema []byte
	Managed      bool
}
