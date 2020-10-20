package helm3

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

type InstallTest struct {
	expectedCommand string
	installStep     InstallStep
}

// // sad hack: not sure how to make a common test main for all my subpackages
func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)

}

func TestMixin_UnmarshalInstallStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/install-input.yaml")
	require.NoError(t, err)

	var action InstallAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Install MySQL", step.Description)
	assert.NotEmpty(t, step.Outputs)
	assert.Equal(t, HelmOutput{"mysql-root-password", "porter-ci-mysql", "mysql-root-password", "", "", "", ""}, step.Outputs[0])
	assert.Equal(t, HelmOutput{"mysql-cluster-ip", "", "", "service", "porter-ci-mysql-service", "default", "{.spec.clusterIP}"}, step.Outputs[2])
	assert.Equal(t, "stable/mysql", step.Chart)
	assert.Equal(t, "0.10.2", step.Version)
	assert.Equal(t, true, step.Replace)
	assert.Equal(t, map[string]string{"mysqlDatabase": "mydb", "mysqlUser": "myuser",
		"livenessProbe.initialDelaySeconds": "30", "persistence.enabled": "true", "controller.nodeSelector.\"beta\\.kubernetes\\.io/os\"": "linux"}, step.Set)
}

func TestMixin_Install(t *testing.T) {
	namespace := "MY-NAMESPACE"
	name := "MYRELEASE"
	chart := "MYCHART"
	version := "1.0.0"
	setArgs := map[string]string{
		"foo":                            "bar",
		"baz":                            "qux",
		"prop2.prop3.\"key1.key2.key3\"": "value2",
		"prop1.prop2.\"key1.key2.key3\".prop1.prop2.\"key1.key2.key3\"": "value1",
	}
	values := []string{
		"/tmp/val1.yaml",
		"/tmp/val2.yaml",
	}

	baseInstall := fmt.Sprintf(`helm3 install %s %s --namespace %s --version %s`, name, chart, namespace, version)
	baseInstallUpSert := fmt.Sprintf(`helm3 upgrade --install %s %s --namespace %s --version %s`, name, chart, namespace, version)
	baseValues := `--values /tmp/val1.yaml --values /tmp/val2.yaml`
	baseSetArgs := `--set baz=qux --set foo=bar --set prop1.prop2."key1\.key2\.key3".prop1.prop2."key1\.key2\.key3"=value1 --set prop2.prop3."key1\.key2\.key3"=value2`

	installTests := []InstallTest{
		{
			expectedCommand: fmt.Sprintf(`%s %s %s`, baseInstall, baseValues, baseSetArgs),
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Step:      Step{Description: "Install Foo"},
					Namespace: namespace,
					Name:      name,
					Chart:     chart,
					Version:   version,
					Set:       setArgs,
					Values:    values,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s`, baseInstall, `--replace`, baseValues, baseSetArgs),
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Step:      Step{Description: "Install Foo"},
					Namespace: namespace,
					Name:      name,
					Chart:     chart,
					Version:   version,
					Set:       setArgs,
					Values:    values,
					Replace:   true,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s`, baseInstall, `--devel`, baseValues, baseSetArgs),
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Step:      Step{Description: "Install Foo"},
					Namespace: namespace,
					Name:      name,
					Chart:     chart,
					Version:   version,
					Set:       setArgs,
					Values:    values,
					Devel:     true,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s`, baseInstallUpSert, `--wait`, baseValues, baseSetArgs),
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Step:      Step{Description: "Install Foo"},
					Namespace: namespace,
					Name:      name,
					Chart:     chart,
					Version:   version,
					Set:       setArgs,
					Values:    values,
					Wait:      true,
					UpSert:    true,
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, installTest := range installTests {
		t.Run(installTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, installTest.expectedCommand)

			action := InstallAction{Steps: []InstallStep{installTest.installStep}}
			b, _ := yaml.Marshal(action)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err := h.Install()

			require.NoError(t, err)
		})
	}
}
