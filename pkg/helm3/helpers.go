package helm3

import (
	"testing"

	"get.porter.sh/porter/pkg/context"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
)

const MockHelmClientVersion string = "v3.6.3"

type TestMixin struct {
	*Mixin
	TestContext *context.TestContext
}

type testKubernetesFactory struct {
}

func (t *testKubernetesFactory) GetClient(configPath string) (kubernetes.Interface, error) {
	return testclient.NewSimpleClientset(), nil
}

// NewTestMixin initializes a mixin test client, with the output buffered, and an in-memory file system.
func NewTestMixin(t *testing.T) *TestMixin {
	c := context.NewTestContext(t)
	m := New()
	m.Context = c.Context
	m.ClientFactory = &testKubernetesFactory{}
	m.HelmClientVersion = MockHelmClientVersion

	return &TestMixin{
		Mixin:       m,
		TestContext: c,
	}
}
