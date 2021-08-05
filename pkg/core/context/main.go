package context

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Context contains every context information gathered during an updatecli run
type Context struct {
	Name       string
	PipelineID string // PipelineID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	Title      string // Title is used for the full pipelin

	Sources    map[string]Source
	Conditions map[string]Condition
	Targets    map[string]Target

	SourcesStageReport    []reports.Stage
	ConditionsStageReport []reports.Stage
	TargetsStageReport    []reports.Stage

	Config *config.Config
}

// Source hold context information gathered during an updatecli run
// that are specific to a source resource
type Source struct {
	Output    string
	Result    string
	Changelog string
}

// Condition hold context information gathered during an updatecli run
// that are specific to a condition resource
type Condition struct {
	Result string
}

// Target hold context information gathered during an updatecli run
// that are specific to a condition resource
type Target struct {
	Result string
}

// Init initialize a updatecli context based on its bind configuration
func (c *Context) Init(config *config.Config) {

	c.Title = config.Title
	c.Name = config.Name
	c.PipelineID = config.PipelineID

	c.Config = config

	// Init context resource size
	c.Sources = make(map[string]Source, len(config.Sources))
	c.Conditions = make(map[string]Condition, len(config.Conditions))
	c.Targets = make(map[string]Target, len(config.Targets))

	// Init sources report
	for id, source := range config.Sources {
		c.SourcesStageReport = append(
			c.SourcesStageReport,
			reports.Stage{
				Name:   source.Name,
				Kind:   source.Kind,
				Result: result.FAILURE,
			})
		c.Sources[id] = Source{
			Result: result.FAILURE,
		}
	}

	// Init conditions report
	for id, condition := range config.Conditions {
		c.ConditionsStageReport = append(
			c.ConditionsStageReport,
			reports.Stage{
				Name:   condition.Name,
				Kind:   condition.Kind,
				Result: result.FAILURE,
			})
		c.Conditions[id] = Condition{
			Result: result.FAILURE,
		}
	}

	// Init target report
	for id, target := range config.Conditions {
		c.TargetsStageReport = append(
			c.TargetsStageReport,
			reports.Stage{
				Name:   target.Name,
				Kind:   target.Kind,
				Result: result.FAILURE,
			})
		c.Targets[id] = Target{
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
