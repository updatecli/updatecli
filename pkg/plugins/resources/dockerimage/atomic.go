package dockerimage

/*
GetAtomicSpec returns an atomic version of the resource spec
The goal is to only have information that can be used to identify
the scope of the resource.
All credentials are absent
*/
func (d DockerImage) GetAtomicSpec() interface{} {
	return d.spec.Atomic()
}
