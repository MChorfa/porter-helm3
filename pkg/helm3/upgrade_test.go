package helm3

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type UpgradeTest struct {
	expectedCommand string
	upgradeStep     UpgradeStep
}

func TestMixin_UnmarshalUpgradeStep(t *testing.T) {
	b, err := os.ReadFile("testdata/upgrade-input.yaml")
	require.NoError(t, err)

	var action UpgradeAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Upgrade MySQL", step.Description)
	assert.NotEmpty(t, step.Outputs)
	assert.Equal(t, HelmOutput{"mysql-root-password", "porter-ci-mysql", "mysql-root-password", "", "", "", ""}, step.Outputs[0])
	assert.Equal(t, HelmOutput{"mysql-cluster-ip", "", "", "service", "porter-ci-mysql-service", "default", "{.spec.clusterIP}"}, step.Outputs[2])
	assert.Equal(t, "stable/mysql", step.Chart)
	assert.Equal(t, "0.10.2", step.Version)
	assert.True(t, step.Wait)
	assert.True(t, step.ResetValues)
	assert.True(t, step.ResetValues)
	assert.Equal(t, map[string]string{"mysqlDatabase": "mydb", "mysqlUser": "myuser",
		"livenessProbe.initialDelaySeconds": "30", "persistence.enabled": "true"}, step.Set)
	assert.Nil(t, step.Atomic)
}

func TestMixin_UnmarshalUpgradeStepAtomicFalse(t *testing.T) {
	b, err := os.ReadFile("testdata/upgrade-input-atomic-false.yaml")
	require.NoError(t, err)

	var action UpgradeAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.False(t, *step.Atomic)
}

func TestMixin_UnmarshalUpgradeStepAtomicTrue(t *testing.T) {
	b, err := os.ReadFile("testdata/upgrade-input-atomic-true.yaml")
	require.NoError(t, err)

	var action UpgradeAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.True(t, *step.Atomic)
}

func TestMixin_UnmarshalUpgradeStepCreateNamespaceFalse(t *testing.T) {
	b, err := os.ReadFile("testdata/upgrade-input-create-namespace-false.yaml")
	require.NoError(t, err)

	var action UpgradeAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.False(t, *step.CreateNamespace)
}

func TestMixin_UnmarshalUpgradeStepCreateNamespaceTrue(t *testing.T) {
	b, err := os.ReadFile("testdata/upgrade-input-create-namespace-true.yaml")
	require.NoError(t, err)

	var action UpgradeAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.True(t, *step.CreateNamespace)
}

func TestMixin_Upgrade(t *testing.T) {
	namespace := "MY-NAMESPACE"
	name := "MY-RELEASE"
	chart := "MY-CHART"
	version := "1.0.0"
	setArgs := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	values := []string{
		"/tmp/val1.yaml",
		"/tmp/val2.yaml",
	}

	baseUpgrade := fmt.Sprintf(`helm3 upgrade --install %s %s --namespace %s --version %s`, name, chart, namespace, version)
	baseNamespaceCreationFlag := `--create-namespace`
	baseValues := `--values /tmp/val1.yaml --values /tmp/val2.yaml`
	baseSetArgs := `--set baz=qux --set foo=bar`
	baseAddFlags := `--atomic`

	valueTrue := true
	valueFalse := false

	upgradeTests := []UpgradeTest{
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s`, baseUpgrade, baseValues, baseAddFlags, baseSetArgs),
			upgradeStep: UpgradeStep{
				UpgradeArguments: UpgradeArguments{
					Step:            Step{Description: "Upgrade Foo"},
					Namespace:       namespace,
					Name:            name,
					Chart:           chart,
					Version:         version,
					Set:             setArgs,
					Values:          values,
					CreateNamespace: &valueFalse,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s %s`, baseUpgrade, `--reset-values`, baseValues, baseAddFlags, baseSetArgs),
			upgradeStep: UpgradeStep{
				UpgradeArguments: UpgradeArguments{
					Step:            Step{Description: "Upgrade Foo"},
					Namespace:       namespace,
					Name:            name,
					Chart:           chart,
					Version:         version,
					Set:             setArgs,
					Values:          values,
					ResetValues:     true,
					CreateNamespace: &valueFalse,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s %s`, baseUpgrade, `--reuse-values`, baseValues, baseAddFlags, baseSetArgs),
			upgradeStep: UpgradeStep{
				UpgradeArguments: UpgradeArguments{
					Step:            Step{Description: "Upgrade Foo"},
					Namespace:       namespace,
					Name:            name,
					Chart:           chart,
					Version:         version,
					Set:             setArgs,
					Values:          values,
					ReuseValues:     true,
					CreateNamespace: &valueFalse,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s %s %s`, baseUpgrade, `--wait`, baseValues, baseAddFlags, baseNamespaceCreationFlag, baseSetArgs),
			upgradeStep: UpgradeStep{
				UpgradeArguments: UpgradeArguments{
					Step:      Step{Description: "Upgrade Foo"},
					Namespace: namespace,
					Name:      name,
					Chart:     chart,
					Version:   version,
					Set:       setArgs,
					Values:    values,
					Wait:      true,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s %s`, baseUpgrade, baseValues, `--timeout 600 --debug`, baseAddFlags, baseSetArgs),
			upgradeStep: UpgradeStep{
				UpgradeArguments: UpgradeArguments{
					Step:            Step{Description: "Upgrade Foo"},
					Namespace:       namespace,
					Name:            name,
					Chart:           chart,
					Version:         version,
					Set:             setArgs,
					Values:          values,
					Timeout:         "600",
					Debug:           true,
					CreateNamespace: &valueFalse,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s`, baseUpgrade, baseValues, baseSetArgs),
			upgradeStep: UpgradeStep{
				UpgradeArguments: UpgradeArguments{
					Step:            Step{Description: "Upgrade Foo"},
					Namespace:       namespace,
					Name:            name,
					Chart:           chart,
					Version:         version,
					Set:             setArgs,
					Values:          values,
					Atomic:          &valueFalse,
					CreateNamespace: &valueFalse,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s %s`, baseUpgrade, baseValues, baseAddFlags, baseNamespaceCreationFlag, baseSetArgs),
			upgradeStep: UpgradeStep{
				UpgradeArguments: UpgradeArguments{
					Step:            Step{Description: "Upgrade Foo"},
					Namespace:       namespace,
					Name:            name,
					Chart:           chart,
					Version:         version,
					Set:             setArgs,
					Values:          values,
					Atomic:          &valueTrue,
					CreateNamespace: &valueTrue,
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, upgradeTest := range upgradeTests {
		t.Run(upgradeTest.expectedCommand, func(t *testing.T) {
			ctx := context.Background()
			os.Setenv(test.ExpectedCommandEnv, upgradeTest.expectedCommand)

			action := UpgradeAction{Steps: []UpgradeStep{upgradeTest.upgradeStep}}
			b, err := yaml.Marshal(action)
			require.NoError(t, err)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err = h.Upgrade(ctx)

			require.NoError(t, err)
		})
	}
}
