package scm

// Scm defines ...
type Scm interface {
	Add(file string)
	Clone() string
	GetDirectory() (directory string)
	Init()
	Push()
	Commit(file, message string)
	Clean()
}