package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var mockRungopass = func(path string) (string, error) {
	return "mocked-value", nil
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name: "data field",
			input: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  password: gopass:mysecret/password
`,
			expectedOutput: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  password: bW9ja2VkLXZhbHVl
`,
		},
		{
			name: "stringData field",
			input: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
stringData:
  username: gopass:mysecret/username
`,
			expectedOutput: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
stringData:
  username: mocked-value
`,
		},
		{
			name: "both data and stringData fields",
			input: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  password: gopass:mysecret/password
stringData:
  username: gopass:mysecret/username
`,
			expectedOutput: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  password: bW9ja2VkLXZhbHVl
stringData:
  username: mocked-value
`,
		},
		{
			name: "no gopass prefix",
			input: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  password: notgopass:mysecret/password
stringData:
  username: notgopass:mysecret/username
`,
			expectedOutput: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  password: notgopass:mysecret/password
stringData:
  username: notgopass:mysecret/username
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := &kio.ByteReadWriter{
				Reader:                bytes.NewBufferString(tt.input),
				Writer:                &bytes.Buffer{},
				KeepReaderAnnotations: false,
			}

			config := &GopassSecret{
				rungopassFunc: mockRungopass,
			}
			process(rw, config)

			output := rw.Writer.(*bytes.Buffer).String()
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}

func TestClearAnnotations(t *testing.T) {
	input := `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
  annotations:
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/function: "true"
    other-annotation: "value"
`
	expectedOutput := `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
  annotations:
    other-annotation: "value"
`

	node, err := yaml.Parse(input)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	err = clearAnnotations(node)
	if err != nil {
		t.Fatalf("failed to clear annotations: %v", err)
	}

	output, err := node.String()
	if err != nil {
		t.Fatalf("failed to convert node to string: %v", err)
	}

	if output != expectedOutput {
		t.Fatalf("expected %s but got %s", expectedOutput, output)
	}
}
