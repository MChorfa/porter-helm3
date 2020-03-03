package helm3

// import (
// 	"get.porter.sh/porter/pkg/exec/builder"
// )

// var _ builder.ExecutableAction = Action{}

// type Action struct {
// 	Steps []Steps // using UnmarshalYAML so that we don't need a custom type per action
// }

// // UnmarshalYAML takes any yaml in this form
// // ACTION:
// // - helm3: ...
// // and puts the steps into the Action.Steps field
// func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	var steps []Steps
// 	results, err := builder.UnmarshalAction(unmarshal, &steps)
// 	if err != nil {
// 		return err
// 	}

// 	// Go doesn't have generics, nothing to see here...
// 	for _, result := range results {
// 		step := result.(*[]Steps)
// 		a.Steps = append(a.Steps, *step...)
// 	}
// 	return nil
// }

// func (a Action) GetSteps() []builder.ExecutableStep {
// 	// Go doesn't have generics, nothing to see here...
// 	steps := make([]builder.ExecutableStep, len(a.Steps))
// 	for i := range a.Steps {
// 		steps[i] = a.Steps[i]
// 	}

// 	return steps
// }

// type Steps struct {
// 	Step `yaml:"helm3"`
// }

// var _ builder.ExecutableStep = Step{}
// var _ builder.StepWithOutputs = Step{}

// type Step struct {
// 	Name        string        `yaml:"name"`
// 	Description string        `yaml:"description"`
// 	Arguments   []string      `yaml:"arguments,omitempty"`
// 	Flags       builder.Flags `yaml:"flags,omitempty"`
// 	Outputs     []Output      `yaml:"outputs,omitempty"`
// }

// func (s Step) GetCommand() string {
// 	return "helm3"
// }

// func (s Step) GetArguments() []string {
// 	return s.Arguments
// }

// func (s Step) GetFlags() builder.Flags {
// 	return s.Flags
// }

// func (s Step) GetOutputs() []builder.Output {
// 	// Go doesn't have generics, nothing to see here...
// 	outputs := make([]builder.Output, len(s.Outputs))
// 	for i := range s.Outputs {
// 		outputs[i] = s.Outputs[i]
// 	}
// 	return outputs
// }

import (
	"get.porter.sh/porter/pkg/exec/builder"
)

var _ builder.ExecutableAction = Action{}

type Action struct {
	Steps []ExecuteSteps // using UnmarshalYAML so that we don't need a custom type per action
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - helm: ...
// and puts the steps into the Action.Steps field
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var steps []ExecuteSteps
	results, err := builder.UnmarshalAction(unmarshal, &steps)
	if err != nil {
		return err
	}

	for _, result := range results {
		step := result.(*[]ExecuteSteps)
		a.Steps = append(a.Steps, *step...)
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

var _ builder.ExecutableStep = ExecuteStep{}

type ExecuteSteps struct {
	ExecuteStep `yaml:"helm3"`
}

type ExecuteStep struct {
	Step        `yaml:",inline"`
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Arguments   []string      `yaml:"arguments,omitempty"`
	Flags       builder.Flags `yaml:"flags,omitempty"`
	Outputs     []Output      `yaml:"outputs,omitempty"`
	Namespace   string        `yaml:"namespace,omitempty"`
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

var _ builder.OutputJsonPath = Output{}
var _ builder.OutputFile = Output{}
var _ builder.OutputRegex = Output{}

type Output struct {
	Name   string `yaml:"name"`
	Secret string `yaml:"secret,omitempty"`
	Key    string `yaml:"key,omitempty"`
	// See https://porter.sh/mixins/exec/#outputs
	// TODO: If your mixin doesn't support these output types, you can remove these and the interface assertions above, and from #/definitions/outputs in schema.json
	JSONPath string `yaml:"JSONPath,omitempty"`
	FilePath string `yaml:"path,omitempty"`
	Regex    string `yaml:"regex,omitempty"`
}

func (o Output) GetName() string {
	return o.Name
}

func (o Output) GetJsonPath() string {
	return o.JSONPath
}

func (o Output) GetFilePath() string {
	return o.FilePath
}

func (o Output) GetRegex() string {
	return o.Regex
}
