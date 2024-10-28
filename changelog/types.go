package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type (
	ResourceType string
	DiffType     string
)

const (
	ResourceKind   ResourceType = "resource"
	DataSourceKind ResourceType = "datasource"

	ChangeTypeAdd    DiffType = "Add"
	ChangeTypeRemove DiffType = "Remove"
	ChangeTypeChange DiffType = "Change"
)

type ItemMap map[ResourceType]map[string]*Item

type Item struct {
	Name        string           `json:"name"`
	Path        string           `json:"path"`
	Description string           `json:"description"`
	ForceNew    bool             `json:"forceNew"`
	Optional    bool             `json:"optional"`
	Sensitive   bool             `json:"sensitive"`
	MaxItems    int              `json:"maxItems"`
	Deprecated  string           `json:"deprecated"`
	Type        schema.ValueType `json:"type"`
	ElemType    schema.ValueType `json:"elemType"`
}

type Diff struct {
	Type         DiffType
	ResourceType ResourceType
	Description  string
	Item         *Item
}

func (c *Diff) String() string {
	// resource name + field name
	path := strings.SplitN(c.Item.Path, ".", 2)

	// e.g.: "Add resource `aiven_project`"
	msg := fmt.Sprintf("%s %s `%s`", c.Type, c.ResourceType, path[0])

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
