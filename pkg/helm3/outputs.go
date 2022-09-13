package helm3

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (m *Mixin) getSecret(ctx context.Context, client kubernetes.Interface, namespace, name, key string) ([]byte, error) {
	if namespace == "" {
		namespace = "default"
	}
	if m.DebugMode {
		fmt.Fprintf(os.Stderr, "Retrieving secret %s/%s and using key %s as an output\n", namespace, name, key)
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("error getting secret %s/%s: %s", namespace, name, err)
	}
	val, ok := secret.Data[key]
	if !ok {
		return nil, fmt.Errorf("couldn't find key %s in secret %s/%s", key, namespace, name)
	}
	return val, nil
}

func (m *Mixin) getOutput(ctx context.Context, resourceType, resourceName, namespace, jsonPath string) ([]byte, error) {
	args := []string{"get", resourceType, resourceName}
	args = append(args, fmt.Sprintf("-o=jsonpath=%s", jsonPath))
	if namespace != "" {
		args = append(args, fmt.Sprintf("--namespace=%s", namespace))
	}
	cmd := m.NewCommand(ctx, "kubectl", args...)
	cmd.Stderr = m.Err
	out, err := cmd.Output()
	if err != nil {
		prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
		return nil, errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
	}
	return out, nil
}

func (m *Mixin) handleOutputs(ctx context.Context, client kubernetes.Interface, namespace string, outputs []HelmOutput) error {
	var outputError error
	//Now get the outputs
	for _, output := range outputs {

		if output.Secret != "" && output.Key != "" {
			// Override namespace if output.Namespace is set
			if output.Namespace != "" {
				namespace = output.Namespace
			}

			val, err := m.getSecret(ctx, client, namespace, output.Secret, output.Key)

			if err != nil {
				return err
			}

			outputError = m.Context.WriteMixinOutputToFile(output.Name, val)
		}

		if output.ResourceType != "" && output.ResourceName != "" && output.JSONPath != "" {
			bytes, err := m.getOutput(ctx,
				output.ResourceType,
				output.ResourceName,
				output.Namespace,
				output.JSONPath,
			)
			if err != nil {
				return err
			}

			outputError = m.Context.WriteMixinOutputToFile(output.Name, bytes)

		}

		if outputError != nil {
			return errors.Wrapf(outputError, "unable to write output '%s'", output.Name)
		}
	}
	return nil
}
