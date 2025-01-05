package ci

import "os"

type Jenkins struct {
}

func (j Jenkins) Name() string {
	return "Jenkins pipeline link"
}

func (j Jenkins) URL() string {
	return os.Getenv("BUILD_URL")
}

func (gha Jenkins) IsDebug() bool {
	return false
}
