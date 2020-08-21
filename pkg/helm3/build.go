package helm3

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/exec/builder"
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// clientVersionConstraint represents the semver constraint for the Helm client version
// Currently, this mixin only supports Helm clients versioned v3.x.x
const clientVersionConstraint string = "^v3.x"

// BuildInput represents stdin passed to the mixin for the build command.
type BuildInput struct {
	Config MixinConfig
}

// MixinConfig represents configuration that can be set on the helm3 mixin in porter.yaml
// mixins:
// - helm3:
// 	  clientVersion: v3.3.0
//	  repositories:
//	    stable:
//		  url: "https://kubernetes-charts.storage.googleapis.com"

type MixinConfig struct {
	ClientVersion string `yaml:"clientVersion,omitempty"`
	Repositories  map[string]Repository
}

type Repository struct {
	URL string `yaml:"url,omitempty"`
}

// Build will generate the necessary Dockerfile lines
// for an invocation image using this mixin
func (m *Mixin) Build() error {

	// Create new Builder.
	var input BuildInput
	err := builder.LoadAction(m.Context, "", func(contents []byte) (interface{}, error) {
		err := yaml.Unmarshal(contents, &input)
		return &input, err
	})
	if err != nil {
		return err
	}

	suppliedClientVersion := input.Config.ClientVersion
	if suppliedClientVersion != "" {
		ok, err := validate(suppliedClientVersion, clientVersionConstraint)
		if err != nil {
			return err
		}
		if !ok {
			return errors.Errorf("supplied clientVersion %q does not meet semver constraint %q",
				suppliedClientVersion, clientVersionConstraint)
		}
		m.HelmClientVersion = suppliedClientVersion
	}

	// Install helm3
	fmt.Fprintf(m.Out, "RUN apt-get update && apt-get install -y curl")
	fmt.Fprintf(m.Out, "\nRUN curl https://get.helm.sh/helm-%s-linux-amd64.tar.gz --output helm3.tar.gz", m.HelmClientVersion)
	fmt.Fprintf(m.Out, "\nRUN tar -xvf helm3.tar.gz")
	fmt.Fprintf(m.Out, "\nRUN mv linux-amd64/helm /usr/local/bin/helm3")

	// Go through repositories
	for name, repo := range input.Config.Repositories {

		commandValue, err := GetAddRepositoryCommand(name, repo.URL)
		if err != nil && m.Debug {
			fmt.Fprintf(m.Err, "DEBUG: addition of repository failed: %s\n", err.Error())
		} else {
			fmt.Fprintf(m.Out, strings.Join(commandValue, " "))
		}
	}
	return nil
}

func GetAddRepositoryCommand(name, url string) (commandValue []string, err error) {

	var commandBuilder []string

	if url == "" {
		return commandBuilder, fmt.Errorf("repository url must be supplied")
	}

	commandBuilder = append(commandBuilder, "\nRUN", "helm3", "repo", "add", name, url)

	return commandBuilder, nil
}

// validate validates that the supplied clientVersion meets the supplied semver constraint
func validate(clientVersion, constraint string) (bool, error) {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, errors.Wrapf(err, "unable to parse version constraint %q", constraint)
	}

	v, err := semver.NewVersion(clientVersion)
	if err != nil {
		return false, errors.Wrapf(err, "supplied client version %q cannot be parsed as semver", clientVersion)
	}

	return c.Check(v), nil
}
