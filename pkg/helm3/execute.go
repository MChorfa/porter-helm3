package helm3

import (
	"context"

	"get.porter.sh/porter/pkg/exec/builder"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func (m *Mixin) loadAction(ctx context.Context) (*Action, error) {
	var action Action
	err := builder.LoadAction(ctx, m.RuntimeConfig, "", func(contents []byte) (interface{}, error) {
		err := yaml.Unmarshal(contents, &action)
		return &action, err
	})
	return &action, err
}

func (m *Mixin) Execute(ctx context.Context) error {
	action, err := m.loadAction(ctx)
	if err != nil {
		return err
	}
	if len(action.Steps) != 1 {
		return errors.Errorf("expected a single step, but got %d", len(action.Steps))
	}
	step := action.Steps[0]

	_, err = builder.ExecuteSingleStepAction(ctx, m.RuntimeConfig, action)
	if err != nil {
		return errors.Wrapf(err, "invocation of action %s failed", action)
	}

	kubeClient, err := m.getKubernetesClient()
	if err != nil {
		return errors.Wrap(err, "couldn't get kubernetes client")
	}

	err = m.handleOutputs(ctx, kubeClient, step.Namespace, step.Outputs)
	return err
}
