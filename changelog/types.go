package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type RootKind string

const (
	ResourceRootKind   RootKind = "resource"
	DataSourceRootKind RootKind = "datasource"
)

type DiffAction string

const (
	AddDiffAction    DiffAction = "Add"
	RemoveDiffAction DiffAction = "Remove"
	ChangeDiffAction DiffAction = "Change"
)

type ItemMap map[RootKind]map[string]*Item

type Item struct {
	Name string   `json:"name"` // e.g. project
	Root string   `json:"root"` // e.g. aiven_project
	Path string   `json:"path"` // e.g. aiven_project.project
	Kind RootKind `json:"kind"` // e.g. resource or datasource

	// Terraform schema fields
	Description string           `json:"description"`
	ForceNew    bool             `json:"forceNew"`
	Optional    bool             `json:"optional"`
	Sensitive   bool             `json:"sensitive"`
	MaxItems    int              `json:"maxItems"`
	Deprecated  string           `json:"deprecated"`
	Type        schema.ValueType `json:"type"`
	ElementType schema.ValueType `json:"elementType"`
}

type Diff struct {
	Action        DiffAction
	Item          *Item
	AlsoAppliesTo *Item // e.g., when the change is same for resource and datasource
	Description   string
}

func (c *Diff) String() string {
	// Often the same diff applies both for resource and datasource
	kinds := make([]string, 0, 2)
	kinds = append(kinds, string(c.Item.Kind))
	if c.AlsoAppliesTo != nil {
		kinds = append(kinds, string(c.AlsoAppliesTo.Kind))
	}

	// e.g.: "Add `aiven_project` resource"
	path := strings.SplitN(c.Item.Path, ".", 2)
	msg := fmt.Sprintf("%s `%s` %s", c.Action, path[0], strings.Join(kinds, " and "))

	// e.g.: "field `project`"
	if len(path) > 1 {
		msg = fmt.Sprintf("%s field `%s`", msg, path[1])
		if len(findEnums(c.Item.Description)) > 0 {
			msg += " (enum)"
		}
	}

	// Adds beta if needed
	if hasBeta(c.Description) {
		msg = fmt.Sprintf("%s _(beta)_", msg)
	}

	// Adds description
	if c.Description != "" {
		msg = fmt.Sprintf("%s: %s", msg, c.Description)
	}

	const maxSize = 120
	return shorten(maxSize, msg)
}
