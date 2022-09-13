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

type InstallAction struct {
	Steps []InstallStep `yaml:"install"`
}

type InstallStep struct {
	InstallArguments `yaml:"helm3"`
}

type InstallArguments struct {
	Step `yaml:",inline"`

	Namespace string            `yaml:"namespace"`
	Name      string            `yaml:"name"`
	Chart     string            `yaml:"chart"`
	Devel     bool              `yaml:"devel"`
	NoHooks   bool              `yaml:"noHooks"`
	Repo      string            `yaml:"repo"`
	Set       map[string]string `yaml:"set"`
	SkipCrds  bool              `yaml:"skipCrds"`
	Password  string            `yaml:"password"`
	Username  string            `yaml:"username"`
	Values    []string          `yaml:"values"`
	Version   string            `yaml:"version"`
	Wait      bool              `yaml:"wait"`
	Timeout   string            `yaml:"timeout"`
	Debug     bool              `yaml:"debug"`
	Atomic    *bool             `yaml:"atomic,omitempty"`
}

func (m *Mixin) Install(ctx context.Context) error {

	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	kubeClient, err := m.getKubernetesClient()
	if err != nil {
		return errors.Wrap(err, "couldn't get kubernetes client")
	}

	var action InstallAction
	err = yaml.Unmarshal(payload, &action)
	if err != nil {
		return err
	}
	if len(action.Steps) != 1 {
		return errors.Errorf("expected a single step, but got %d", len(action.Steps))
	}
	step := action.Steps[0]

	cmd := m.NewCommand(ctx, "helm3")

	cmd.Args = append(cmd.Args, "upgrade", "--install", step.Name, step.Chart)

	if step.Namespace != "" {
		cmd.Args = append(cmd.Args, "--namespace", step.Namespace)
	}

	if step.Version != "" {
		cmd.Args = append(cmd.Args, "--version", step.Version)
	}

	if step.Wait {
		cmd.Args = append(cmd.Args, "--wait")
	}

	if step.Devel {
		cmd.Args = append(cmd.Args, "--devel")
	}

	for _, v := range step.Values {
		cmd.Args = append(cmd.Args, "--values", v)
	}

	if step.SkipCrds {
		cmd.Args = append(cmd.Args, "--skip-crds")
	}

	if step.NoHooks {
		cmd.Args = append(cmd.Args, "--no-hooks")
	}

	if step.Repo != "" && step.Username != "" && step.Password != "" {
		cmd.Args = append(cmd.Args, "--repo", step.Repo, "--username", step.Username, "--password", step.Password)
	}

	if step.Timeout != "" {
		cmd.Args = append(cmd.Args, "--timeout", step.Timeout)
	}
	if step.Debug {
		cmd.Args = append(cmd.Args, "--debug")
	}

	if step.Atomic == nil || *step.Atomic {
		// This will ensure the installation process deletes the installation on failure.
		cmd.Args = append(cmd.Args, "--atomic")
	}
	// This will ensure the creation of the release namespace if not present.
	cmd.Args = append(cmd.Args, "--create-namespace")
	// Set values
	cmd.Args = HandleSettingChartValuesForInstall(step, cmd)

	cmd.Stdout = m.Out
	cmd.Stderr = m.Err

	// format the command with all arguments
	prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
	fmt.Fprintln(m.Out, prettyCmd)

	// Here where really the command get executed
	err = cmd.Start()
	// Exit on error
	if err != nil {
		return fmt.Errorf("could not execute command, %s: %s", prettyCmd, err)
	}
	err = cmd.Wait()
	// Exit on error
	if err != nil {
		return err
	}
	err = m.handleOutputs(ctx, kubeClient, step.Namespace, step.Outputs)
	return err
}

// Prepare set arguments
func HandleSettingChartValuesForInstall(step InstallStep, cmd *exec.Cmd) []string {
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
