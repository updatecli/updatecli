package jenkins

/*
GetAtomicSpec returns an atomic version of the resource spec
The goal is to only have information that can be used to identify
the scope of the resource.
All credentials are absent
*/
func (j Jenkins) GetAtomicSpec() interface{} {
	return j.spec.Atomic()
}
