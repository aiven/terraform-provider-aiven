package main

import (
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

// readHandlerPackages maps operation IDs to handler packages using the config.yaml
// of the go-client-codegen version in go.mod, the way the client generator does.
func readHandlerPackages() (map[OperationID]string, error) {
	out, err := osexec.Command("go", "list", "-m", "-f", "{{.Dir}}", avnGenPackage).Output()
	if err != nil {
		return nil, fmt.Errorf("could not locate %s: %w", avnGenPackage, err)
	}

	data, err := os.ReadFile(filepath.Join(strings.TrimSpace(string(out)), "config.yaml"))
	if err != nil {
		return nil, err
	}

	var config map[string][]OperationID
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not parse go-client-codegen config.yaml: %w", err)
	}

	packages := make(map[OperationID]string)
	for name, ids := range config {
		for _, id := range ids {
			packages[id] = avnGenHandlerPackage + strings.ToLower(name)
		}
	}
	return packages, nil
}

// setHandlerPackages sets Operation.Package for every API operation.
func setHandlerPackages(definitions []*Definition, packages map[OperationID]string) error {
	for _, def := range definitions {
		for _, op := range def.Operations {
			if op.ID == "" {
				continue
			}
			pkg, ok := packages[op.ID]
			if !ok {
				return fmt.Errorf("%s: operation %s not found in go-client-codegen", def.fileName, op.ID)
			}
			op.Package = pkg
		}
	}
	return nil
}

func (o Operations) packages() []string {
	return lo.Uniq(lo.FilterMap(o, func(op *Operation, _ int) (string, bool) {
		return op.Package, op.Package != ""
	}))
}
