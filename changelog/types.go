package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type RootType string

const (
	ResourceRootType   RootType = "resource"
	DataSourceRootType RootType = "datasource"
)

type DiffAction string

const (
	AddDiffAction    DiffAction = "Add"
	RemoveDiffAction DiffAction = "Remove"
	ChangeDiffAction DiffAction = "Change"
)

type ItemMap map[RootType]map[string]*Item

type Item struct {
	Path string `json:"path"` // e.g. aiven_project.project
	Name string `json:"name"` // e.g. project

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
	Action      DiffAction
	RootType    RootType
	Description string
	Item        *Item
}

func (c *Diff) String() string {
	// resource name + field name
	path := strings.SplitN(c.Item.Path, ".", 2)

	// e.g.: "Add `aiven_project` resource"
	msg := fmt.Sprintf("%s `%s` %s", c.Action, path[0], c.RootType)

	// e.g.: "field `project`"
	if len(path) > 1 {
		msg = fmt.Sprintf("%s field `%s`", msg, path[1])
	}

	// Adds beta if needed
	if hasBeta(c.Description) {
		msg = fmt.Sprintf("%s _(beta)_", msg)
	}

	// Adds description
	const maxSize = 120

	msg += ": "
	msg += shorten(maxSize-len(msg), c.Description)
	return msg
}
