package chart

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/olblak/updateCli/pkg/scm"
)

// Condition check if a specific chart version exist
func (c *Chart) Condition(source string) (bool, error) {

	if c.Version != "" {
		fmt.Printf("Version %v, already defined from configuration file\n", c.Version)
	} else {
		c.Version = source
	}
	URL := fmt.Sprintf("%s/index.yaml", c.URL)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return false, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return false, err
	}

	index, err := loadIndex(body)

	if err != nil {
		return false, err
	}

	message := ""
	if c.Version != "" {
		message = fmt.Sprintf(" for version '%s'", c.Version)
	}

	if index.Has(c.Name, c.Version) {
		fmt.Printf("\u2714 Helm Chart '%s' is available on %s%s\n", c.Name, c.URL, message)
		return true, nil
	}

	fmt.Printf("\u2717 Helm Chart '%s' isn't available on %s%s\n", c.Name, c.URL, message)
	return false, nil

}

// ConditionFromSCM returns an error because it's not supported
func (c *Chart) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for Helm chart condition, aborting")
}
