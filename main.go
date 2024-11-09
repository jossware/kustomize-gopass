package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type GopassSecret struct {
	rungopassFunc func(string) (string, error)
}

const gopassPrefix = "gopass:"

func main() {
	process(&kio.ByteReadWriter{}, &GopassSecret{
		rungopassFunc: rungopass,
	})
}

func process(byteReadWriter *kio.ByteReadWriter, config *GopassSecret) {
	// function that will be passed to the kustomize framework execute function
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for _, item := range items {
			if item.GetKind() == "Secret" && item.GetApiVersion() == "v1" {
				if err := clearAnnotations(item); err != nil {
					return items, fmt.Errorf("clearing annotations: %w", err)
				}
				if err := config.gopass(item); err != nil {
					return items, fmt.Errorf("transforming gopass keys: %w", err)
				}
			}
		}

		return items, nil
	}

	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}

	if err := framework.Execute(p, byteReadWriter); err != nil {
		log.Printf("error running gopass secret function: %v", err)
		os.Exit(1)
	}
}

func (gs *GopassSecret) gopass(item *yaml.RNode) error {
	fields := map[string]bool{
		"data":       true,
		"stringData": false,
	}

	for field, b64encode := range fields {
		data := item.Field(field)
		if data == nil {
			continue
		}

		// field should be a `MapNode`. if not, it's not a valid Secret and
		// we will fail with `wrong node kind`
		err := data.Value.VisitFields(func(node *yaml.MapNode) error {
			key := node.Key.YNode().Value
			value := node.Value.YNode().Value

			if strings.HasPrefix(value, gopassPrefix) {
				newValue, err := gs.rungopassFunc(strings.TrimPrefix(value, gopassPrefix))
				if err != nil {
					return fmt.Errorf("getting gopass secret from %s for %s: %w", value, key, err)
				}
				if b64encode {
					newValue = base64.StdEncoding.EncodeToString([]byte(newValue))
				}
				node.Value.YNode().Value = newValue
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
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
	annosToClear := []string{"config.kubernetes.io/local-config", "config.kubernetes.io/function"}

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
