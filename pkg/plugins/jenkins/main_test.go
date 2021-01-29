package jenkins

import (
	"fmt"
	"strings"
	"testing"
)

type Data struct {
	version     string
	err         error
	releaseType string
}

type DataSet []Data

var (
	dataset DataSet = DataSet{
		{
			version:     "2.275",
			releaseType: WEEKLY,
		},
		{
			version:     "2.249.3",
			releaseType: STABLE,
		},
		{
			version:     "2",
			err:         fmt.Errorf("Version 2 contains 1 component(s) which doesn't correspond to any valid release type"),
			releaseType: WRONG,
		},
		{
			version:     "2.249.3.4",
			err:         fmt.Errorf("Version 2.249.3.4 contains 4 component(s) which doesn't correspond to any valid release type"),
			releaseType: WRONG,
		},
		{
			version:     "2.249.3-rc",
			err:         fmt.Errorf("In version '2.249.3-rc', component '3-rc' is not a valid integer"),
			releaseType: WRONG,
		},
	}
)

func TestReleaseType(t *testing.T) {

	for _, data := range dataset {
		got, err := ReleaseType(data.version)
		if err != nil {
			if data.err == nil {
				t.Error(fmt.Errorf("Version '%v' expected the error:\n%v",
					data.version,
					err))
			}

			if strings.Compare(err.Error(), data.err.Error()) != 0 {
				t.Error(fmt.Errorf("For version '%v' expected error:\n%v\ngot\n%v",
					data.version,
					data.err,
					err))
			}
		}
		if got != data.releaseType {
			t.Error(fmt.Errorf("Version validity for %v expected to be %v but got %v ",
				data.version,
				data.releaseType,
				got))
		}

	}
}
