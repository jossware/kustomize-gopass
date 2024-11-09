package main

import (
	"log"
	"os"

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

type GopassSecret struct {
	Data map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
}

func process(byteReadWriter *kio.ByteReadWriter, runAsCommand bool) {
	config := new(GopassSecret)
	log.Printf("Processing secret data, config: %+v", config)

	//function that will be passed to the kustomize framework execute function
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for _, item := range items {
			log.Printf("item: %s", item.GetKind())
		}

		secret, err := yaml.Parse(SECRET_TEMPLATE)
		if err != nil {
			return items, err
		}
		items = append(items, secret)
		return items, nil
	}

	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}

	if !runAsCommand {
		if error := framework.Execute(p, byteReadWriter); error != nil {
			panic(error)
		}
	}

	if runAsCommand {
		cmd := command.Build(p, command.StandaloneDisabled, false)
		//automatically generates the Dockerfile for us
		command.AddGenerateDockerfile(cmd)
		if err := cmd.Execute(); err != nil {
			os.Exit(1)
		}
	}
}

func main() {
	runAsCommand := false
	byteReadWriter := &kio.ByteReadWriter{}
	process(byteReadWriter, runAsCommand)
}
