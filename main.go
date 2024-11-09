package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const SECRET_TEMPLATE = `
apiVersion: v1
kind: Secret
metadata:
 name: testing
data:
 secret: value`

type GopassSecret struct{}

var annosToClear = []string{"config.kubernetes.io/local-config", "config.kubernetes.io/function"}

func gopass(item *yaml.RNode) error {
	// TODO: stringData
	data := item.Field("data")
	if data == nil {
		return nil // Skip if no metadata field
	}

	// Assuming you have an RNode that's a map
	// TODO: what happens if it is not?
	err := data.Value.VisitFields(func(node *yaml.MapNode) error {
		key := node.Key.YNode().Value
		value := node.Value.YNode().Value

		if strings.HasPrefix(value, "gopass:") {
			newValue, err := rungopass(strings.TrimPrefix(value, "gopass:"))
			if err != nil {
				return fmt.Errorf("getting gopass secret from %s for %s: %w", value, key, err)
			}
			node.Value.YNode().Value = newValue
		}
		return nil
	})

	return err
}

func rungopass(path string) (string, error) {
	cmd := exec.Command("gopass", "show", "-o", path)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("running \"gopass show -o %s\": %w", path, err)
	}
	return string(output), nil
}

func clearAnnotations(item *yaml.RNode) error {
	// Get the metadata field
	metadata := item.Field("metadata")
	if metadata == nil {
		return nil // Skip if no metadata field
	}

	// Get the annotations field
	annotations := metadata.Value.Field("annotations")
	if annotations == nil {
		return nil // Skip if no annotations field
	}

	// Remove the specific annotations
	for _, anno := range annosToClear {
		err := annotations.Value.PipeE(yaml.Clear(anno))
		if err != nil {
			return fmt.Errorf("failed to remove annotation %q: %v", anno, err)
		}
	}

	// Remove annotations field if empty
	if len(annotations.Value.YNode().Content) == 0 {
		metadata.Value.PipeE(yaml.Clear("annotations"))
	}

	return nil
}

func process(byteReadWriter *kio.ByteReadWriter, runAsCommand bool) {
	config := new(GopassSecret)
	log.Printf("Processing secret data, config: %+v", config)

	//function that will be passed to the kustomize framework execute function
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for _, item := range items {
			log.Printf("Processing item: %s", item.GetKind())
			if item.GetKind() == "Secret" && item.GetApiVersion() == "v1" {
				if err := clearAnnotations(item); err != nil {
					return items, fmt.Errorf("clearing annotations: %w", err)
				}
				if err := gopass(item); err != nil {
					return items, fmt.Errorf("transforming gopass keys: %w", err)
				}
			}
		}

		return items, nil
	}

	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}

	if !runAsCommand {
		if err := framework.Execute(p, byteReadWriter); err != nil {
			log.Printf("error running gopass secret function: %v", err)
			os.Exit(1)
		}
	}

	if runAsCommand {
		cmd := command.Build(p, command.StandaloneDisabled, false)
		//automatically generates the Dockerfile for us
		command.AddGenerateDockerfile(cmd)
		if err := cmd.Execute(); err != nil {
			log.Printf("error running gopass secret function: %v", err)
			os.Exit(1)
		}
	}
}

func main() {
	runAsCommand := false
	byteReadWriter := &kio.ByteReadWriter{}
	process(byteReadWriter, runAsCommand)
}
