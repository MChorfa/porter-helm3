package helm3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type UninstallAction struct {
	Steps []UninstallStep `yaml:"uninstall"`
}

// UninstallStep represents the structure of an Uninstall action
type UninstallStep struct {
	UninstallArguments `yaml:"helm3"`
}

// UninstallArguments are the arguments available for the Uninstall action
type UninstallArguments struct {
	Step      `yaml:",inline"`
	Namespace string   `yaml:"namespace,omitempty"`
	Releases  []string `yaml:"releases"`
	NoHooks   bool     `yaml:"noHooks"`
	Wait      bool     `yaml:"wait"`
	Timeout   string   `yaml:"timeout"`
	Debug     bool     `yaml:"debug"`
}

// Uninstall deletes a provided set of Helm releases, supplying optional flags/params
func (m *Mixin) Uninstall(ctx context.Context) error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var action UninstallAction
	err = yaml.Unmarshal(payload, &action)
	if err != nil {
		return err
	}
	if len(action.Steps) != 1 {
		return errors.Errorf("expected a single step, but got %d", len(action.Steps))
	}
	step := action.Steps[0]

	// Delete each release one at a time, because helm stops on first error
	// This gives us more fine-grained error recovery and handling
	var result error
	for _, release := range step.Releases {
		err = m.delete(ctx, release, step.Namespace, step.NoHooks, step.Wait, step.Timeout, step.Debug)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (m *Mixin) delete(ctx context.Context, release string, namespace string, noHooks bool, wait bool, timeout string, debug bool) error {
	cmd := m.NewCommand(ctx, "helm3", "uninstall")

	cmd.Args = append(cmd.Args, release)

	if namespace != "" {
		cmd.Args = append(cmd.Args, "--namespace", namespace)
	}

	if noHooks {
		cmd.Args = append(cmd.Args, "--no-hooks")
	}

	if wait {
		cmd.Args = append(cmd.Args, "--wait")
	}

	if timeout != "" {
		cmd.Args = append(cmd.Args, "--timeout", timeout)
	}

	if debug {
		cmd.Args = append(cmd.Args, "--debug")
	}
	output := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(m.Out, output)
	cmd.Stderr = io.MultiWriter(m.Err, output)

	prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
	fmt.Fprintln(m.Out, prettyCmd)

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("could not execute command, %s: %s", prettyCmd, err)
	}
	err = cmd.Wait()
	if err != nil {
		// Gracefully handle the error being a release not loaded or found
		outputBuffer := strings.ToLower(output.String())
		if strings.Contains(outputBuffer, fmt.Sprintf(`uninstall: release not loaded: %q`, strings.ToLower(release))) ||
			strings.Contains(outputBuffer, fmt.Sprintf(`the release named %q is already deleted`, strings.ToLower(release))) ||
			strings.Contains(outputBuffer, fmt.Sprintf(`release: %q not found`, strings.ToLower(release))) ||
			strings.Contains(outputBuffer, "release not loaded") ||
			strings.Contains(outputBuffer, "already deleted") ||
			strings.Contains(outputBuffer, "release: not found") ||
			strings.Contains(outputBuffer, "not found") {
			return nil
		}
		return err
	}

	return nil
}
