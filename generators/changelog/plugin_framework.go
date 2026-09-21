package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
)

// fromPluginFrameworkProvider processes a Plugin Framework provider
func fromPluginFrameworkProvider(p *plugin.AivenProvider) (ItemMap, error) {
	var (
		ctx   = context.Background()
		items = ItemMap{
			ResourceRootKind:   make(map[string]*Item),
			DataSourceRootKind: make(map[string]*Item),
		}
	)

	for _, resourceFunc := range p.Resources(ctx) {
		res := resourceFunc()

		var (
			metaResp   resource.MetadataResponse
			schemaResp resource.SchemaResponse
		)

		res.Metadata(ctx, resource.MetadataRequest{}, &metaResp)
		res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

		rootItem := &Item{
			Name:        metaResp.TypeName,
			Root:        metaResp.TypeName,
			Path:        metaResp.TypeName,
			Kind:        ResourceRootKind,
			Description: schemaResp.Schema.MarkdownDescription,
			Deprecated:  schemaResp.Schema.DeprecationMessage,
			Type:        schema.TypeList,
		}
		items[ResourceRootKind][rootItem.Path] = rootItem

		for _, item := range walkPluginSchema(schemaResp.Schema, rootItem) {
			items[ResourceRootKind][item.Path] = item
		}
	}

	for _, datasourceFunc := range p.DataSources(ctx) {
		ds := datasourceFunc()

		var (
			metaResp   datasource.MetadataResponse
			schemaResp datasource.SchemaResponse
		)

		ds.Metadata(ctx, datasource.MetadataRequest{}, &metaResp)
		ds.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)

		rootItem := &Item{
			Name:        metaResp.TypeName,
			Root:        metaResp.TypeName,
			Path:        metaResp.TypeName,
			Kind:        DataSourceRootKind,
			Description: schemaResp.Schema.MarkdownDescription,
			Deprecated:  schemaResp.Schema.DeprecationMessage,
			Type:        schema.TypeList,
		}
		items[DataSourceRootKind][rootItem.Path] = rootItem

		for _, item := range walkPluginSchema(schemaResp.Schema, rootItem) {
			items[DataSourceRootKind][item.Path] = item
		}
	}

	return items, nil
}

// walkPluginSchema walks a schema (resource or data source) and returns a list of discovered items
func walkPluginSchema(s any, parent *Item) []*Item {
	members := make(map[string]any)

	switch sc := s.(type) {
	case resourceschema.Schema:
		collectMembers(members, sc.Attributes)
		collectMembers(members, sc.Blocks)
	case datasourceschema.Schema:
		collectMembers(members, sc.Attributes)
		collectMembers(members, sc.Blocks)
	default:
		log.Panicf("unsupported schema type: %T", sc)
	}

	return walkMembers(members, parent)
}

// walkMembers walks the given attributes and blocks and returns a list of discovered items
func walkMembers(members map[string]any, parent *Item) []*Item {
	var items []*Item
	for name, member := range members {
		items = append(items, walkMember(name, member, parent)...)
	}
	return items
}

// walkMember walks an attribute or a block and its nested elements
func walkMember(name string, member any, parent *Item) []*Item {
	var description, deprecated string

	if d, ok := member.(interface{ GetMarkdownDescription() string }); ok {
		description = d.GetMarkdownDescription()
	}

	if d, ok := member.(interface{ GetDeprecationMessage() string }); ok {
		deprecated = d.GetDeprecationMessage()
	}

	item := &Item{
		Name:        name,
		Root:        parent.Root,
		Path:        fmt.Sprintf("%s.%s", parent.Path, name),
		Kind:        parent.Kind,
		Description: description,
		Deprecated:  deprecated,
		Type:        pluginFrameworkTypeToSDKType(member),
		ForceNew:    requiresReplace(member),
	}

	// Blocks implement neither of these, so they stay at their zero value.
	if o, ok := member.(interface{ IsOptional() bool }); ok {
		item.Optional = o.IsOptional()
	}
	if s, ok := member.(interface{ IsSensitive() bool }); ok {
		item.Sensitive = s.IsSensitive()
	}

	return append([]*Item{item}, walkMembers(nestedMembers(member), item)...)
}

// collectMembers adds the given attributes or blocks to members.
// Attribute and block maps are typed per schema package, so they are collapsed into a single
// map to let walkMember handle both without caring where they came from.
func collectMembers[T any](members map[string]any, src map[string]T) {
	for name, member := range src {
		members[name] = member
	}
}

// nestedMembers returns the attributes and blocks nested in the given attribute or block.
//
// Nesting is exposed by the framework through its internal fwschema package, which cannot be
// imported here, so the exported concrete types are matched instead. Every nesting type the
// plugin generator can emit must be listed, otherwise its children are skipped and their
// changes never reach the changelog.
func nestedMembers(member any) map[string]any {
	members := make(map[string]any)

	switch m := member.(type) {
	case resourceschema.ListNestedAttribute:
		collectMembers(members, m.NestedObject.Attributes)
	case resourceschema.SetNestedAttribute:
		collectMembers(members, m.NestedObject.Attributes)
	case resourceschema.MapNestedAttribute:
		collectMembers(members, m.NestedObject.Attributes)
	case resourceschema.SingleNestedAttribute:
		collectMembers(members, m.Attributes)
	case resourceschema.ListNestedBlock:
		collectMembers(members, m.NestedObject.Attributes)
		collectMembers(members, m.NestedObject.Blocks)
	case resourceschema.SetNestedBlock:
		collectMembers(members, m.NestedObject.Attributes)
		collectMembers(members, m.NestedObject.Blocks)
	case resourceschema.SingleNestedBlock:
		collectMembers(members, m.Attributes)
		collectMembers(members, m.Blocks)

	case datasourceschema.ListNestedAttribute:
		collectMembers(members, m.NestedObject.Attributes)
	case datasourceschema.SetNestedAttribute:
		collectMembers(members, m.NestedObject.Attributes)
	case datasourceschema.MapNestedAttribute:
		collectMembers(members, m.NestedObject.Attributes)
	case datasourceschema.SingleNestedAttribute:
		collectMembers(members, m.Attributes)
	case datasourceschema.ListNestedBlock:
		collectMembers(members, m.NestedObject.Attributes)
		collectMembers(members, m.NestedObject.Blocks)
	case datasourceschema.SetNestedBlock:
		collectMembers(members, m.NestedObject.Attributes)
		collectMembers(members, m.NestedObject.Blocks)
	case datasourceschema.SingleNestedBlock:
		collectMembers(members, m.Attributes)
		collectMembers(members, m.Blocks)
	}

	return members
}

// requiresReplaceDescription is the description stringplanmodifier.RequiresReplace and its
// siblings report. It is the only public marker distinguishing them from other plan modifiers.
const requiresReplaceDescription = "Terraform will destroy and recreate the resource"

// requiresReplace reports whether the attribute or block forces recreation on change.
// PlanModifiers is typed per value kind ([]planmodifier.String, []planmodifier.Bool, ...),
// so the field is read reflectively rather than through a case per attribute type.
func requiresReplace(member any) bool {
	v := reflect.ValueOf(member)
	if v.Kind() != reflect.Struct {
		return false
	}

	modifiers := v.FieldByName("PlanModifiers")
	if !modifiers.IsValid() || modifiers.Kind() != reflect.Slice {
		return false
	}

	ctx := context.Background()
	for i := range modifiers.Len() {
		d, ok := modifiers.Index(i).Interface().(interface {
			Description(context.Context) string
		})
		if ok && strings.Contains(d.Description(ctx), requiresReplaceDescription) {
			return true
		}
	}

	return false
}

// pluginFrameworkTypeToSDKType converts a plugin framework attribute or block type to its SDKv2 equivalent
func pluginFrameworkTypeToSDKType(member any) schema.ValueType {
	switch member.(type) {
	case resourceschema.StringAttribute, datasourceschema.StringAttribute:
		return schema.TypeString
	case resourceschema.BoolAttribute, datasourceschema.BoolAttribute:
		return schema.TypeBool
	case resourceschema.Int64Attribute, datasourceschema.Int64Attribute:
		return schema.TypeInt
	case resourceschema.Float64Attribute, datasourceschema.Float64Attribute:
		return schema.TypeFloat
	case resourceschema.MapAttribute, datasourceschema.MapAttribute,
		resourceschema.MapNestedAttribute, datasourceschema.MapNestedAttribute:
		return schema.TypeMap
	case resourceschema.SetAttribute, datasourceschema.SetAttribute,
		resourceschema.SetNestedAttribute, datasourceschema.SetNestedAttribute,
		resourceschema.SetNestedBlock, datasourceschema.SetNestedBlock:
		return schema.TypeSet
	// treated as TypeList
	case resourceschema.ListAttribute, datasourceschema.ListAttribute,
		resourceschema.ListNestedAttribute, datasourceschema.ListNestedAttribute,
		resourceschema.SingleNestedAttribute, datasourceschema.SingleNestedAttribute,
		resourceschema.ListNestedBlock, datasourceschema.ListNestedBlock,
		resourceschema.SingleNestedBlock, datasourceschema.SingleNestedBlock:
		return schema.TypeList
	default:
		log.Panicf("unsupported attribute type: %T", member) // should not happen

		return schema.TypeString
	}
}
