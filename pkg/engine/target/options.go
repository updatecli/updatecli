package target

// Options hold target parameters
type Options struct {
	Commit bool
	Push   bool
	Clean  bool
}

// Init Options with default values
func (o *Options) Init() {
	o.Commit = true
	o.Push = true
	o.Clean = true
}
