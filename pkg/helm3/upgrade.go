package helm3

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
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

	Namespace   string            `yaml:"namespace"`
	Name        string            `yaml:"name"`
	Chart       string            `yaml:"chart"`
	Version     string            `yaml:"version"`
	Set         map[string]string `yaml:"set"`
	Values      []string          `yaml:"values"`
	Wait        bool              `yaml:"wait"`
	ResetValues bool              `yaml:"resetValues"`
	ReuseValues bool              `yaml:"reuseValues"`
}

// Upgrade issues a helm upgrade command for a release using the provided UpgradeArguments
func (m *Mixin) Upgrade() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	kubeClient, err := m.getKubernetesClient("/root/.kube/config")
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

	cmd := m.NewCommand("helm3", "upgrade", "--install", step.Name, step.Chart)

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

	err = m.handleOutputs(kubeClient, step.Namespace, step.Outputs)
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
		//Hack unitl helm introduce `--set-literal` for complex keys
		// see https://github.com/helm/helm/issues/4030
		// TODO : Fix this later upon `--set-literal` introduction
		var complexKey bool
		keySequences := strings.Split(k, ".")
		escapedKeys := make([]string, 0, len(keySequences))
		for _, ks := range keySequences {
			// Start Complex key
			if strings.HasPrefix(ks, "\"") {
				escapedKeys = append(escapedKeys, ks+"\\")
				complexKey = true
				// Still in the middle of complex key
			} else if !strings.HasPrefix(ks, "\"") && !strings.HasSuffix(ks, "\"") && complexKey {
				escapedKeys = append(escapedKeys, ks+"\\")
				// Reach the end of complex key
			} else if strings.HasSuffix(ks, "\"") && complexKey {
				escapedKeys = append(escapedKeys, ks)
				complexKey = false
				// Do nothing
			} else {
				escapedKeys = append(escapedKeys, ks)
			}
		}

		cmd.Args = append(cmd.Args, "--set", fmt.Sprintf("%s=%s", strings.Join(escapedKeys, "."), step.Set[k]))
	}
	return cmd.Args
}
