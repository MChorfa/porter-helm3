package helm3

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/exec/builder"
	yaml "gopkg.in/yaml.v2"
)

// This is an example. Replace the following with whatever steps are needed to
// install required components into
// const dockerfileLines = `RUN apt-get update && \
// apt-get install gnupg apt-transport-https lsb-release software-properties-common -y && \
// echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ stretch main" | \
//    tee /etc/apt/sources.list.d/azure-cli.list && \
// apt-key --keyring /etc/apt/trusted.gpg.d/Microsoft.gpg adv \
// 	--keyserver packages.microsoft.com \
// 	--recv-keys BC528686B50D79E339D3721CEB3E94ADBE1229CF && \
// apt-get update && apt-get install azure-cli
// `

// BuildInput represents stdin passed to the mixin for the build command.
type BuildInput struct {
	Config MixinConfig
}

// MixinConfig represents configuration that can be set on the helm mixin in porter.yaml
// mixins:
// - helm:
//	  repositories:
//	    stable:
//		  url: "https://kubernetes-charts.storage.googleapis.com"
//		  cafile: "path/to/cafile"
//		  certfile: "path/to/certfile"
//		  keyfile: "path/to/keyfile"
//		  username: "username"
//		  password: "password"
type MixinConfig struct {
	Repositories map[string]Repository
}

type Repository struct {
	URL      string `yaml:"url,omitempty"`
	Cafile   string `yaml:"cafile,omitempty"`
	Certfile string `yaml:"certfile,omitempty"`
	Keyfile  string `yaml:"keyfile,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
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
	// Make sure kubectl is available
	// fmt.Fprintln(m.Out, `RUN apt-get update && apt-get install -y apt-transport-https curl`)
	// fmt.Fprintln(m.Out, `RUN curl https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/linux/amd64/kubectl --output kubectl`)
	// fmt.Fprintln(m.Out, `RUN mv kubectl /usr/local/bin`)
	// fmt.Fprintln(m.Out, `RUN chmod a+x /usr/local/bin/kubectl`)

	// Install helm3
	fmt.Fprintln(m.Out, `RUN apt-get update && apt-get install -y curl`)
	fmt.Fprintln(m.Out, `RUN curl https://get.helm.sh/helm-v3.1.2-linux-amd64.tar.gz --output helm3.tar.gz`)
	fmt.Fprintln(m.Out, `RUN tar -xvf helm3.tar.gz`)
	fmt.Fprintln(m.Out, `RUN mv linux-amd64/helm /usr/local/bin/helm3`)

	// Go through repositories
	for name, repo := range input.Config.Repositories {

		commandValue, err := GetAddRepositoryCommand(name, repo.URL, repo.Cafile, repo.Certfile, repo.Keyfile, repo.Username, repo.Password)
		if err != nil && m.Debug {
			fmt.Fprintf(m.Err, "DEBUG: addition of repository failed: %s\n", err.Error())
		} else {
			fmt.Fprintf(m.Out, strings.Join(commandValue, " "))
		}
	}
	return nil
}

func GetAddRepositoryCommand(name, url, cafile, certfile, keyfile, username, password string) (commandValue []string, err error) {

	var commandBuilder []string

	if url == "" {
		return commandBuilder, fmt.Errorf("repository url must be supplied")
	}

	commandBuilder = append(commandBuilder, "\nRUN", "helm", "repo", "add", name, url)

	if certfile != "" && keyfile != "" {
		commandBuilder = append(commandBuilder, "--cert-file", certfile, "--key-file", keyfile)
	}
	if cafile != "" {
		commandBuilder = append(commandBuilder, "--ca-file", cafile)
	}
	if username != "" && password != "" {
		commandBuilder = append(commandBuilder, "--username", username, "--password", password)
	}

	return commandBuilder, nil
}
