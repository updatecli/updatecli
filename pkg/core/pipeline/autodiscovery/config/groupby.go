package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type GroupBy string

const (
	GROUPEBYALL        GroupBy = "all"
	GROUPEBYINDIVIDUAL GroupBy = "individual"
)

func (g GroupBy) Validate() error {
	valid := false
	for _, groupBy := range []GroupBy{
		GROUPEBYALL,
		GROUPEBYINDIVIDUAL,
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
			GROUPEBYALL,
			GROUPEBYINDIVIDUAL)

		logrus.Errorln(err)
		return err

	}
	return nil
}
