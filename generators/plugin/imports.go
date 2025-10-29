package main

import (
	"fmt"
	"strings"
)

// entityType TF entity type
type entityType string

const (
	resourceType   entityType = "resource"
	datasourceType entityType = "datasource"
)

func (e entityType) isResource() bool {
	return e == resourceType
}

func (e entityType) Title() string {
	if e == resourceType {
		return "Resource"
	}

	// In camelcase
	return "DataSource"
}

func (e entityType) Import(importString entityImportType) string {
	return fmt.Sprintf(string(importString), e)
}

func boolEntity(isResource bool) entityType {
	if isResource {
		return resourceType
	}
	return datasourceType
}

// Generic untyped imports
const (
	projectPackagePrefix  = "github.com/aiven/terraform-provider-aiven"
	attrPackage           = "github.com/hashicorp/terraform-plugin-framework/attr"
	diagPackage           = "github.com/hashicorp/terraform-plugin-framework/diag"
	typesPackage          = "github.com/hashicorp/terraform-plugin-framework/types"
	pathPackage           = "github.com/hashicorp/terraform-plugin-framework/path"
	validatorPackage      = "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	utilPackage           = "github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	adapterPackage        = "github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	legacyTimeoutsPackage = "github.com/aiven/terraform-provider-aiven/internal/plugin/legacytimeouts"
	avnGenPackage         = "github.com/aiven/go-client-codegen"
	errMsgPackage         = "github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func getUntypedImports() []string {
	return []string{
		attrPackage,
		diagPackage,
		typesPackage,
		pathPackage,
		validatorPackage,
		utilPackage,
		adapterPackage,
		legacyTimeoutsPackage,
		errMsgPackage,
	}
}

// Read strings to be formatted with the entity type
type entityImportType string

const (
	entityPackage          entityImportType = "github.com/hashicorp/terraform-plugin-framework/%s"
	schemaPackageFmt       entityImportType = "github.com/hashicorp/terraform-plugin-framework/%s/schema"
	planmodifierPackageFmt entityImportType = "github.com/hashicorp/terraform-plugin-framework/%s/schema/planmodifier"
	timeoutsPackageFmt     entityImportType = "github.com/hashicorp/terraform-plugin-framework-timeouts/%s/timeouts"
)

func entityImport(isResource bool, importString entityImportType) string {
	return fmt.Sprintf(string(importString), boolEntity(isResource))
}

// typingMapping Terraform internal types mapping
func typingMapping() map[SchemaType]string {
	return map[SchemaType]string{
		SchemaTypeBoolean:      "Bool",
		SchemaTypeNumber:       "Float64",
		SchemaTypeInteger:      "Int64",
		SchemaTypeObject:       "List", // We use List type for objects, because it is compatible with SDKv2
		SchemaTypeArray:        "Set",
		SchemaTypeArrayOrdered: "List",
		SchemaTypeString:       "String",
	}
}

// typedImport import that depends on type (string
type typedImport string

const (
	planmodifierTypedImport typedImport = "github.com/hashicorp/terraform-plugin-framework/resource/schema/%splanmodifier"
	validatorTypedImport    typedImport = "github.com/hashicorp/terraform-plugin-framework-validators/%svalidator"
	defaultsTypedImport     typedImport = "github.com/hashicorp/terraform-plugin-framework/resource/schema/%sdefault"
)

func getTypedImports() []typedImport {
	return []typedImport{
		planmodifierTypedImport,
		validatorTypedImport,
		defaultsTypedImport,
	}
}

func getTypedImport(kind SchemaType, imp typedImport) string {
	suffix := strings.ToLower(typingMapping()[kind])
	return fmt.Sprintf(string(imp), suffix)
}
