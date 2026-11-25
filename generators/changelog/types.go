package main

import (
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
	AlsoAppliesTo *Item // when the change is the same for resource and datasource
	Description   string
}
