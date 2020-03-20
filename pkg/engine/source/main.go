package source

// Source defines how a value is retrieved from a specific source
type Source struct {
	Kind    string
	Output  string
	Prefix  string
	Postfix string
	Spec    interface{}
}
