package helm3

import (
	"fmt"
	"os/exec"
	"sort"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type InstallAction struct {
	Steps []InstallStep `yaml:"install"`
}

type InstallStep struct {
	InstallArguments `yaml:"helm3"`
}

type InstallArguments struct {
	Step `yaml:",inline"`

	Namespace    string                `yaml:"namespace"`
	Name         string                `yaml:"name"`
	Chart        string                `yaml:"chart"`
	Version      string                `yaml:"version"`
	Replace      bool                  `yaml:"replace"`
	Set          map[string]string     `yaml:"set"`
	Values       []string              `yaml:"values"`
	Devel        bool                  `yaml:"devel`
	UpSert       bool                  `yaml:"devel`
	Wait         bool                  `yaml:"wait"`
	RegistryAuth RegistryAuthArguments `yaml:"registryAuth"`
}

type RegistryAuthArguments struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (step *InstallStep) GetChart() string {
	return step.Chart
}
func (step *InstallStep) GetRegistryUsername() string {
	return step.RegistryAuth.Username
}
func (step *InstallStep) GetRegistryPassword() string {
	return step.RegistryAuth.Password
}
func (step *InstallStep) GetOptionalVersion() string {
	return step.Version
}

func (m *Mixin) Install() error {

	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	kubeClient, err := m.getKubernetesClient("/root/.kube/config")
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

	isOciInstall := IsOciInstall(&step)

	if isOciInstall {
		fmt.Fprintln(m.Out, "OCI install detected.")

		LoginToOciRegistryIfNecessary(&step, m)
		PullChartFromOciRegistry(&step, m)
		newChartName, err := ExportOciChartToTempPath(&step, m)

		if err != nil {
			return err
		}
		step.Chart = newChartName
	} else {
		fmt.Fprintln(m.Out, "No OCI install detected.")
	}

	cmd := m.NewCommand("helm3")

	if step.UpSert {
		cmd.Args = append(cmd.Args, "upgrade", "--install", step.Name, step.Chart)

	} else {
		cmd.Args = append(cmd.Args, "install", step.Name, step.Chart)
	}

	if step.Namespace != "" {
		cmd.Args = append(cmd.Args, "--namespace", step.Namespace)
	}

	if step.Version != "" {
		cmd.Args = append(cmd.Args, "--version", step.Version)
	}

	if step.Replace {
		cmd.Args = append(cmd.Args, "--replace")
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

	cmd.Args = HandleSettingChartValuesForInstall(step, cmd)

	err = m.RunCmd(cmd, true)

	// Exit on error
	if err != nil {
		return err
	}
	err = m.handleOutputs(kubeClient, step.Namespace, step.Outputs)
	if err != nil {
		return err
	}

	if isOciInstall {
		err = RemoveLocalOciExport(&step, m)
	}
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
