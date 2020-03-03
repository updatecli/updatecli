package scm

// Scm is an interface in from of source controle manager like git or github
type Scm interface {
	Add(file string)
	Clone() string
	GetDirectory() (directory string)
	Init(version string)
	Push()
	Commit(file, message string)
	Clean()
}
