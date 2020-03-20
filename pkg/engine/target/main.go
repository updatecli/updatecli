package target

// Target defines which file need to be updated based on source output
type Target struct {
	Name       string
	Kind       string
	Spec       interface{}
	Repository interface{}
}
