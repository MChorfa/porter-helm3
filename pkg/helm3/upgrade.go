package helm3

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type UpgradeAction struct {
	Steps []UpgradeStep `yaml:"upgrade"`
}

// UpgradeStep represents the structure of an Upgrade step
type UpgradeStep struct {
	UpgradeArguments `yaml:"helm3"`
}

// UpgradeArguments represent the arguments available to the Upgrade step
type UpgradeArguments struct {
	Step `yaml:",inline"`

	Namespace       string            `yaml:"namespace"`
	Name            string            `yaml:"name"`
	Chart           string            `yaml:"chart"`
	Version         string            `yaml:"version"`
	NoHooks         bool              `yaml:"nohooks"`
	Set             map[string]string `yaml:"set"`
	Values          []string          `yaml:"values"`
	Wait            bool              `yaml:"wait"`
	ResetValues     bool              `yaml:"resetValues"`
	ReuseValues     bool              `yaml:"reuseValues"`
	Repo            string            `yaml:"repo"`
	SkipCrds        bool              `yaml:"skipCrds"`
	Password        string            `yaml:"password"`
	Username        string            `yaml:"username"`
	Timeout         string            `yaml:"timeout"`
	Debug           bool              `yaml:"debug"`
	Atomic          *bool             `yaml:"atomic,omitempty"`
	CreateNamespace *bool             `yaml:"createNamespace,omitempty"`
}

// Upgrade issues a helm upgrade command for a release using the provided UpgradeArguments
func (m *Mixin) Upgrade(ctx context.Context) error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	kubeClient, err := m.getKubernetesClient()
	if err != nil {
		return errors.Wrap(err, "couldn't get kubernetes client")
	}

	var action UpgradeAction
	err = yaml.Unmarshal(payload, &action)
	if err != nil {
		return err
	}
	if len(action.Steps) != 1 {
		return errors.Errorf("expected a single step, but got %d", len(action.Steps))
	}
	step := action.Steps[0]

	cmd := m.NewCommand(ctx, "helm3", "upgrade", "--install", step.Name, step.Chart)

	if step.Namespace != "" {
		cmd.Args = append(cmd.Args, "--namespace", step.Namespace)
	}

	if step.Version != "" {
		cmd.Args = append(cmd.Args, "--version", step.Version)
	}

	if step.ResetValues {
		cmd.Args = append(cmd.Args, "--reset-values")
	}

	if step.ReuseValues {
		cmd.Args = append(cmd.Args, "--reuse-values")
	}

	if step.Wait {
		cmd.Args = append(cmd.Args, "--wait")
	}

	for _, v := range step.Values {
		cmd.Args = append(cmd.Args, "--values", v)
	}

	if step.Timeout != "" {
		cmd.Args = append(cmd.Args, "--timeout", step.Timeout)
	}

	if step.Debug {
		cmd.Args = append(cmd.Args, "--debug")
	}

	if step.Atomic == nil || *step.Atomic {
		// This will upgrade process rolls back changes made in case of failed upgrade.
		cmd.Args = append(cmd.Args, "--atomic")
	}

	if step.CreateNamespace == nil || *step.CreateNamespace {
		// This will ensure the creation of the release namespace if not present.
		cmd.Args = append(cmd.Args, "--create-namespace")
	}

	cmd.Args = HandleSettingChartValuesForUpgrade(step, cmd)

	cmd.Stdout = m.Out
	cmd.Stderr = m.Err

	prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
	fmt.Fprintln(m.Out, prettyCmd)

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not execute command, %s: %s", prettyCmd, err)
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}

	err = m.handleOutputs(ctx, kubeClient, step.Namespace, step.Outputs)
	return err
}

// Prepare set arguments
func HandleSettingChartValuesForUpgrade(step UpgradeStep, cmd *exec.Cmd) []string {
	// sort the set consistently
	setKeys := make([]string, 0, len(step.Set))
	for k := range step.Set {

		setKeys = append(setKeys, k)
	}
	sort.Strings(setKeys)

	for _, k := range setKeys {
		cmd.Args = append(cmd.Args, "--set", fmt.Sprintf("%s=%s", k, step.Set[k]))
	}
	return cmd.Args
}
