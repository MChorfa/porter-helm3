package helm3

import (
	"get.porter.sh/porter/pkg/exec/builder"
)

var _ builder.ExecutableAction = Action{}
var _ builder.BuildableAction = Action{}
var _ builder.ExecutableStep = ExecuteStep{}

type Action struct {
	// Name of the action: install, upgrade, invoke, uninstall.
	Name  string
	Steps []ExecuteSteps // using UnmarshalYAML so that we don't need a custom type per action
}

// MarshalYAML converts the action back to a YAML representation
// install:
//   exec:
//     ...
//   helm3:
//     ...
func (a Action) MarshalYAML() (interface{}, error) {
	return map[string]interface{}{a.Name: a.Steps}, nil
}

// MakeSteps builds a slice of Steps for data to be unmarshaled into.
func (a Action) MakeSteps() interface{} {
	return &[]ExecuteSteps{}
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - helm3: ...
// and puts the steps into the Action.Steps field
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	results, err := builder.UnmarshalAction(unmarshal, a)
	if err != nil {
		return err
	}

	for actionName, action := range results {
		a.Name = actionName
		for _, result := range action {
			step := result.(*[]ExecuteSteps)
			a.Steps = append(a.Steps, *step...)
		}
		break // There is only 1 action
	}
	return nil
}

func (a Action) GetSteps() []builder.ExecutableStep {
	steps := make([]builder.ExecutableStep, len(a.Steps))
	for i := range a.Steps {
		steps[i] = a.Steps[i]
	}

	return steps
}

type ExecuteSteps struct {
	ExecuteStep `yaml:"helm3"`
}

type ExecuteStep struct {
	Step      `yaml:",inline"`
	Namespace string        `yaml:"namespace,omitempty"`
	Arguments []string      `yaml:"arguments,omitempty"`
	Flags     builder.Flags `yaml:"flags,omitempty"`
}

func (s ExecuteStep) GetWorkingDir() string {
	return "."
}

func (s ExecuteStep) GetCommand() string {
	return "helm3"
}

func (s ExecuteStep) GetArguments() []string {
	return s.Arguments
}

func (s ExecuteStep) GetFlags() builder.Flags {
	return s.Flags
}
