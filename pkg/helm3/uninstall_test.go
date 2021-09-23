package helm3

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

type UninstallTest struct {
	expectedCommand string
	uninstallStep   UninstallStep
}

func TestMixin_UnmarshalUninstallStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/uninstall-input.yaml")
	require.NoError(t, err)

	var action UninstallAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Uninstall MySQL", step.Description)
	assert.Equal(t, []string{"porter-ci-mysql"}, step.Releases)
	assert.Equal(t, true, step.Wait)
	assert.Equal(t, true, step.NoHooks)
}

func TestMixin_Uninstall(t *testing.T) {
	releases := []string{
		"foo",
	}
	namespace := "my-namespace"
	noHooks := true
	wait := true

	uninstallTests := []UninstallTest{
		{
			expectedCommand: "helm3 uninstall foo",
			uninstallStep: UninstallStep{
				UninstallArguments: UninstallArguments{
					Step:     Step{Description: "Uninstall Foo"},
					Releases: releases,
				},
			},
		},
		{
			expectedCommand: "helm3 uninstall foo --namespace my-namespace",
			uninstallStep: UninstallStep{
				UninstallArguments: UninstallArguments{
					Step:      Step{Description: "Uninstall Foo"},
					Releases:  releases,
					Namespace: namespace,
				},
			},
		},
		{
			expectedCommand: "helm3 uninstall foo --namespace my-namespace --no-hooks --wait",
			uninstallStep: UninstallStep{
				UninstallArguments: UninstallArguments{
					Step:      Step{Description: "Uninstall Foo"},
					Releases:  releases,
					Namespace: namespace,
					NoHooks:   noHooks,
					Wait:      wait,
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, uninstallTest := range uninstallTests {
		t.Run(uninstallTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, uninstallTest.expectedCommand)

			action := UninstallAction{Steps: []UninstallStep{uninstallTest.uninstallStep}}
			b, _ := yaml.Marshal(action)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err := h.Uninstall()

			require.NoError(t, err)
		})
	}
}
