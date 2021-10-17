package context

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/engine/condition"
	"github.com/updatecli/updatecli/pkg/core/engine/source"
	"github.com/updatecli/updatecli/pkg/core/engine/target"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Context contains every context information gathered during an updatecli run
type Context struct {
	Name       string
	PipelineID string // PipelineID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	Title      string // Title is used for the full pipelin

	Sources    map[string]source.Source
	Conditions map[string]condition.Condition
	Targets    map[string]target.Target

	SourcesStageReport    []reports.Stage
	ConditionsStageReport []reports.Stage
	TargetsStageReport    []reports.Stage

	Config *config.Config
}

// Init initialize a updatecli context based on its bind configuration
func (c *Context) Init(config *config.Config) {

	c.Title = config.Title
	c.Name = config.Name
	c.PipelineID = config.PipelineID

	c.Config = config

	// Init context resource size
	c.Sources = make(map[string]source.Source, len(config.Sources))
	c.Conditions = make(map[string]condition.Condition, len(config.Conditions))
	c.Targets = make(map[string]target.Target, len(config.Targets))

	// Init sources report
	for id := range config.Sources {
		c.SourcesStageReport = append(
			c.SourcesStageReport,
			reports.Stage{
				Name:   config.Sources[id].Name,
				Kind:   config.Sources[id].Kind,
				Result: result.FAILURE,
			})

		// Init Sources[id]
		c.Sources[id] = source.Source{
			Spec:   config.Sources[id],
			Result: result.FAILURE,
		}

	}

	// Init conditions report
	for id := range config.Conditions {
		c.ConditionsStageReport = append(
			c.ConditionsStageReport,
			reports.Stage{
				Name:   config.Conditions[id].Name,
				Kind:   config.Conditions[id].Kind,
				Result: result.FAILURE,
			})

		c.Conditions[id] = condition.Condition{
			Spec:   config.Conditions[id],
			Result: result.FAILURE,
		}
	}

	// Init target report
	for id := range config.Targets {
		c.TargetsStageReport = append(
			c.TargetsStageReport,
			reports.Stage{
				Name:   config.Targets[id].Name,
				Kind:   config.Targets[id].Kind,
				Result: result.FAILURE,
			})

		c.Targets[id] = target.Target{
			Spec:   config.Targets[id],
			Result: result.FAILURE,
		}
	}

}

func (c *Context) String() string {

	result := fmt.Sprintf("%q: %q\n", "Name", c.Name)
	result = result + fmt.Sprintf("%q: %q\n", "Title", c.Title)
	result = result + fmt.Sprintf("%q: %q\n", "PipelineID", c.PipelineID)

	result = result + fmt.Sprintf("%q:\n", "Sources")
	for key, value := range c.Sources {
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Changelog", value.Changelog)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Output", value.Output)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}
	result = result + fmt.Sprintf("%q:\n", "Conditions")
	for key, value := range c.Conditions {
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}
	result = result + fmt.Sprintf("%q:\n", "Targets")
	for key, value := range c.Targets {
		result = result + fmt.Sprintf("\t%q: %q\n", key, value)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}

	return result
}
