package autodiscovery

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type GroupBy string

const (
	GROUPBYALL        GroupBy = "all"
	GROUPBYINDIVIDUAL GroupBy = "individual"
)

func (g GroupBy) Validate() error {
	valid := false
	for _, groupBy := range []GroupBy{
		GROUPBYALL,
		GROUPBYINDIVIDUAL,
		"",
	} {
		if len(g) == len(groupBy) {
			valid = true
			break
		}
	}

	if !valid {
		err := fmt.Errorf("autodiscovery key, 'groupby' is wrongly set to %q, and must be one of [%q,%q,%q]",
			g,
			"",
			GROUPBYALL,
			GROUPBYINDIVIDUAL)

		logrus.Errorln(err)
		return err

	}
	return nil
}
