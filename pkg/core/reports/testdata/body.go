package body

type Body struct {
	Name      string
	Pipelines []Pipeline
}

type Pipeline struct {
	Name       string
	Conditions []ConditionReport
	Targets    []TargetReport
}

type ConditionReport struct {
	Name   string
	Result string
}

type TargetReport struct {
	Name      string
	Changelog string
	Change    string
}
