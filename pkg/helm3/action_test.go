package helm3

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestMixin_UnmarshalStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/step-input.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)

	// step := action.Steps[0]
	// assert.NotEmpty(t, step.Description)
	// assert.Equal([], step.Outputs)
	// assert.Equal(t, Output{Name: "postgresql-root-password", Secret: "porter-ci-postgresql"}, step.Outputs[0])

	// require.Len(t, step.Arguments, 1)
	// assert.Equal(t, "man-e-faces", step.Arguments[0])

	// require.Len(t, step.Flags, 1)
	// assert.Equal(t, builder.NewFlag("species", "human"), step.Flags[0])
}
