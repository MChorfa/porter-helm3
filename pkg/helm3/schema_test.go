package helm3

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml" // We are not using go-yaml because of serialization problems with jsonschema, don't use this library elsewhere
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestMixin_PrintSchema(t *testing.T) {
	m := NewTestMixin(t)

	err := m.PrintSchema()
	require.NoError(t, err)

	gotSchema := m.TestContext.GetOutput()

	wantSchema, err := ioutil.ReadFile("schema/schema.json")
	require.NoError(t, err)

	assert.Equal(t, string(wantSchema), gotSchema)
}

func TestMixin_ValidateSchema(t *testing.T) {
	m := NewTestMixin(t)

	// Load the mixin schema
	schemaB := m.GetSchema()
	schemaLoader := gojsonschema.NewStringLoader(schemaB)

	// This validates that your schema.json is filled in properly
	testcases := []struct {
		name      string
		file      string
		wantError string
	}{
		{"install", "testdata/install-input.yaml", ""},
		{"invalid property", "testdata/invalid-input.yaml", "Additional property args is not allowed"},
		{"install", "testdata/upgrade-input.yaml", ""},
		{"invalid property", "testdata/invalid-input.yaml", "Additional property args is not allowed"},
		{"install", "testdata/uninstall-input.yaml", ""},
		{"invalid property", "testdata/invalid-input.yaml", "Additional property args is not allowed"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the mixin input as a go dump
			mixinInputB, err := ioutil.ReadFile(tc.file)
			require.NoError(t, err)
			mixinInputMap := make(map[string]interface{})
			err = yaml.Unmarshal(mixinInputB, &mixinInputMap)
			require.NoError(t, err)
			mixinInputLoader := gojsonschema.NewGoLoader(mixinInputMap)

			// Validate the manifest against the schema
			result, err := gojsonschema.Validate(schemaLoader, mixinInputLoader)
			require.NoError(t, err)

			if tc.wantError == "" {
				assert.True(t, result.Valid())
				assert.Empty(t, result.Errors())
			} else {
				assert.False(t, result.Valid())
				assert.Contains(t, fmt.Sprintf("%v", result.Errors()), tc.wantError)
			}
		})
	}
}
