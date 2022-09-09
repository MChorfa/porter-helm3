package helm3

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/exec/builder"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestMixin_UnmarshalExecuteStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/execute-input.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "MySQL Status", step.Description)
	assert.Equal(t, []string{"status", "mysql"}, step.Arguments)
	wantFlags := builder.Flags{
		builder.Flag{
			Name:   "o",
			Values: []string{"yaml"},
		},
	}
	assert.Equal(t, wantFlags, step.Flags)
	wantOutputs := []HelmOutput{
		{
			Name:   "mysql-root-password",
			Secret: "porter-ci-mysql",
			Key:    "mysql-root-password",
		},
	}
	assert.Equal(t, wantOutputs, step.Outputs)
}

func TestMixin_UnmarshalExecuteLoginRegistryStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/execute-input-login-registry-private.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Login to OCI registry", step.Description)
	assert.Equal(t, []string{"registry", "login", "https://charts.private.com"}, step.Arguments)
	wantFlags := builder.Flags{
		builder.Flag{Name: "p", Values: []string{"mypass"}},
		builder.Flag{Name: "u", Values: []string{"myuser"}},
	}

	sort.Sort(step.Flags)
	assert.EqualValues(t, wantFlags, step.Flags)
}

func TestMixin_UnmarshalExecuteLoginRegistryInsecureStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/execute-input-login-registry-insecure.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Login to OCI registry", step.Description)
	assert.Equal(t, []string{"registry", "login", "localhost:5000", "--insecure"}, step.Arguments)
	wantFlags := builder.Flags{
		builder.Flag{Name: "p", Values: []string{"mypass"}},
		builder.Flag{Name: "u", Values: []string{"myuser"}},
	}

	assert.EqualValues(t, wantFlags, step.Flags)
}

func TestMixin_Execute_Login_Registry(t *testing.T) {
	ctx := context.Background()

	defer os.Unsetenv(test.ExpectedCommandEnv)
	os.Setenv(test.ExpectedCommandEnv, "helm3 registry login localhost:5000 --insecure -p mypass -u myuser")

	executeAction := Action{
		Steps: []ExecuteSteps{
			{
				ExecuteStep: ExecuteStep{
					Arguments: []string{
						"registry",
						"login",
						"localhost:5000",
						"--insecure",
					},
					Flags: builder.Flags{
						{
							Name: "u",
							Values: []string{
								"myuser",
							},
						},
						{
							Name: "p",
							Values: []string{
								"mypass",
							},
						},
					},
				},
			},
		},
	}

	b, _ := yaml.Marshal(executeAction)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	err := h.Execute(ctx)
	require.NoError(t, err)
}

func TestMixin_Execute(t *testing.T) {
	ctx := context.Background()

	defer os.Unsetenv(test.ExpectedCommandEnv)
	os.Setenv(test.ExpectedCommandEnv, "helm3 status mysql -o yaml")

	executeAction := Action{
		Steps: []ExecuteSteps{
			{
				ExecuteStep: ExecuteStep{
					Arguments: []string{
						"status",
						"mysql",
					},
					Flags: builder.Flags{
						{
							Name: "o",
							Values: []string{
								"yaml",
							},
						},
					},
				},
			},
		},
	}

	b, _ := yaml.Marshal(executeAction)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	err := h.Execute(ctx)
	require.NoError(t, err)
}
