package pullrequest

/*
GetAtomicSpec returns an atomic version of the resource spec
The goal is to only have information that can be used to identify
the scope of the resource.
All credentials are absent
*/
func (g Gitea) GetAtomicSpec() interface{} {
	a := g.spec.Atomic()

	if g.Owner != "" {
		a.Owner = g.Owner
	}
	if g.Repository != "" {
		a.Repository = g.Owner
	}

	if g.SourceBranch != "" {
		a.SourceBranch = g.SourceBranch
	}

	if g.TargetBranch != "" {
		a.TargetBranch = g.TargetBranch
	}

	return a
}
