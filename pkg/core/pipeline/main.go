package pipeline

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/engine/condition"
	"github.com/updatecli/updatecli/pkg/core/engine/source"
	"github.com/updatecli/updatecli/pkg/core/engine/target"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Pipeline represent an updatecli run for a specific configuration
type Pipeline struct {
	Name  string // Name defines a pipeline name, used to improve human visualization
	ID    string // ID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	Title string // Title is used for the full pipelin

	Sources    map[string]source.Source
	Conditions map[string]condition.Condition
	Targets    map[string]target.Target

	SourcesStageReport    []reports.Stage
	ConditionsStageReport []reports.Stage
	TargetsStageReport    []reports.Stage

	Config *config.Config
}

// Init initialize an updatecli context based on its configuration
func (p *Pipeline) Init(config *config.Config) {

	p.Title = config.Title
	p.Name = config.Name
	p.ID = config.PipelineID

	p.Config = config

	// Init context resource size
	p.Sources = make(map[string]source.Source, len(config.Sources))
	p.Conditions = make(map[string]condition.Condition, len(config.Conditions))
	p.Targets = make(map[string]target.Target, len(config.Targets))

	// Init sources report
	for id := range config.Sources {
		p.SourcesStageReport = append(
			p.SourcesStageReport,
			reports.Stage{
				Name:   config.Sources[id].Name,
				Kind:   config.Sources[id].Kind,
				Result: result.FAILURE,
			})

		// Init Sources[id]
		p.Sources[id] = source.Source{
			Config: config.Sources[id],
			Result: result.FAILURE,
		}

	}

	// Init conditions report
	for id := range config.Conditions {
		p.ConditionsStageReport = append(
			p.ConditionsStageReport,
			reports.Stage{
				Name:   config.Conditions[id].Name,
				Kind:   config.Conditions[id].Kind,
				Result: result.FAILURE,
			})

		p.Conditions[id] = condition.Condition{
			Config: config.Conditions[id],
			Result: result.FAILURE,
		}
	}

	// Init target report
	for id := range config.Targets {
		p.TargetsStageReport = append(
			p.TargetsStageReport,
			reports.Stage{
				Name:   config.Targets[id].Name,
				Kind:   config.Targets[id].Kind,
				Result: result.FAILURE,
			})

		p.Targets[id] = target.Target{
			Config: config.Targets[id],
			Result: result.FAILURE,
		}
	}

}

func (p *Pipeline) String() string {

	result := fmt.Sprintf("%q: %q\n", "Name", p.Name)
	result = result + fmt.Sprintf("%q: %q\n", "Title", p.Title)
	result = result + fmt.Sprintf("%q: %q\n", "ID", p.ID)

	result = result + fmt.Sprintf("%q:\n", "Sources")
	for key, value := range p.Sources {
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Changelog", value.Changelog)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Output", value.Output)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}
	result = result + fmt.Sprintf("%q:\n", "Conditions")
	for key, value := range p.Conditions {
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}
	result = result + fmt.Sprintf("%q:\n", "Targets")
	for key, value := range p.Targets {
		result = result + fmt.Sprintf("\t%q: %q\n", key, value)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}

	return result
}
