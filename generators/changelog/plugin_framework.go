package main

import (
	"context"
	"fmt"
	"log"

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
	var items []*Item

	switch sc := s.(type) {
	case resourceschema.Schema:
		for name, attr := range sc.GetAttributes() {
			items = append(items, walkAttribute(name, attr, parent)...)
		}

		for name, block := range sc.GetBlocks() {
			items = append(items, walkBlock(name, block, parent)...)
		}
	case datasourceschema.Schema:
		for name, attr := range sc.GetAttributes() {
			items = append(items, walkAttribute(name, attr, parent)...)
		}

		for name, block := range sc.GetBlocks() {
			items = append(items, walkBlock(name, block, parent)...)
		}
	default:
		log.Panicf("unsupported schema type: %T", sc)
	}

	return items
}

// walkAttribute walks an attribute and its nested elements
func walkAttribute(name string, attr any, parent *Item) []*Item {
	var description, deprecated string

	if d, ok := attr.(interface{ GetMarkdownDescription() string }); ok {
		description = d.GetMarkdownDescription()
	}

	if d, ok := attr.(interface{ GetDeprecationMessage() string }); ok {
		deprecated = d.GetDeprecationMessage()
	}

	item := &Item{
		Name:        name,
		Root:        parent.Root,
		Path:        fmt.Sprintf("%s.%s", parent.Path, name),
		Kind:        parent.Kind,
		Description: description,
		Deprecated:  deprecated,
		Type:        pluginFrameworkTypeToSDKType(attr),
	}

	if o, ok := attr.(interface{ IsOptional() bool }); ok {
		item.Optional = o.IsOptional()
	}
	if s, ok := attr.(interface{ IsSensitive() bool }); ok {
		item.Sensitive = s.IsSensitive()
	}
	if f, ok := attr.(interface{ RequiresReplace() bool }); ok {
		item.ForceNew = f.RequiresReplace()
	}
	if m, ok := attr.(interface{ GetMaxItems() int64 }); ok {
		item.MaxItems = int(m.GetMaxItems())
	}

	items := []*Item{item}

	// walk nested object
	if n, ok := attr.(interface{ GetNestedObject() any }); ok {
		nestedObj := n.GetNestedObject()
		if a, ok := nestedObj.(interface{ GetAttributes() any }); ok {
			attrs := a.GetAttributes()
			switch typedAttrs := attrs.(type) {
			case map[string]resourceschema.Attribute:
				for nestedName, nestedAttr := range typedAttrs {
					items = append(items, walkAttribute(nestedName, nestedAttr, item)...)
				}
			case map[string]datasourceschema.Attribute:
				for nestedName, nestedAttr := range typedAttrs {
					items = append(items, walkAttribute(nestedName, nestedAttr, item)...)
				}
			}
		}
	}

	return items
}

// walkBlock walks a block and its nested elements
func walkBlock(name string, block any, parent *Item) []*Item {
	var description, deprecated string
	if d, ok := block.(interface{ GetMarkdownDescription() string }); ok {
		description = d.GetMarkdownDescription()
	}
	if d, ok := block.(interface{ GetDeprecationMessage() string }); ok {
		deprecated = d.GetDeprecationMessage()
	}

	item := &Item{
		Name:        name,
		Root:        parent.Root,
		Path:        fmt.Sprintf("%s.%s", parent.Path, name),
		Kind:        parent.Kind,
		Description: description,
		Deprecated:  deprecated,
		Type:        schema.TypeList, // treat blocks as lists
	}

	items := []*Item{item}

	if n, ok := block.(interface{ GetNestedObject() any }); ok {
		nestedObj := n.GetNestedObject()
		if a, ok := nestedObj.(interface{ GetAttributes() any }); ok {
			attrs := a.GetAttributes()
			switch typedAttrs := attrs.(type) {
			case map[string]resourceschema.Attribute:
				for nestedName, nestedAttr := range typedAttrs {
					items = append(items, walkAttribute(nestedName, nestedAttr, item)...)
				}
			case map[string]datasourceschema.Attribute:
				for nestedName, nestedAttr := range typedAttrs {
					items = append(items, walkAttribute(nestedName, nestedAttr, item)...)
				}
			}
		}
	}

	return items
}

// pluginFrameworkTypeToSDKType converts a plugin framework attribute type to its SDKv2 equivalent
func pluginFrameworkTypeToSDKType(attr any) schema.ValueType {
	switch attr.(type) {
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
		resourceschema.SetNestedAttribute, datasourceschema.SetNestedAttribute:
		return schema.TypeSet
	// treated as TypeList
	case resourceschema.ListAttribute, datasourceschema.ListAttribute,
		resourceschema.ListNestedAttribute, datasourceschema.ListNestedAttribute,
		resourceschema.SingleNestedAttribute, datasourceschema.SingleNestedAttribute:
		return schema.TypeList
	default:
		log.Panicf("unsupported attribute type: %T", attr) // should not happen

		return schema.TypeString
	}
}
