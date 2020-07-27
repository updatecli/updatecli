package reports

import (
	"github.com/olblak/updateCli/pkg/config"
)

// Report contains a list of Rules
type Report struct {
	Name       string
	Result     string
	Source     Stage
	Conditions []Stage
	Targets    []Stage
}

// Update report based on latest information
func (r *Report) Update(config *config.Config) {

	r.Source.Kind = config.Source.Kind
	r.Source.Name = config.Source.Name
	r.Source.Result = config.Source.Result

	i := 0
	for _, condition := range config.Conditions {
		c := &r.Conditions[i]
		c.Name = condition.Name
		c.Kind = condition.Kind
		c.Result = condition.Result
		i++
	}

	i = 0
	for _, target := range config.Targets {
		t := &r.Targets[i]
		t.Name = target.Name
		t.Kind = target.Kind
		t.Result = target.Result
		i++
	}

}