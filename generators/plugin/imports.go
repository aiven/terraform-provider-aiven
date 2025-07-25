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

func boolEntity(isResource bool) entityType {
	if isResource {
		return resourceType
	}
	return datasourceType
}

// Generic untyped imports
const (
	attrPackage      = "github.com/hashicorp/terraform-plugin-framework/attr"
	diagPackage      = "github.com/hashicorp/terraform-plugin-framework/diag"
	typesPackage     = "github.com/hashicorp/terraform-plugin-framework/types"
	validatorPackage = "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	utilPackage      = "github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	adapterPackage   = "github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func getUntypedImports() []string {
	return []string{
		attrPackage,
		diagPackage,
		typesPackage,
		validatorPackage,
		utilPackage,
		adapterPackage,
	}
}

// Read strings to be formatted with the entity type
type entityImportType string

const (
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
		SchemaTypeBoolean: "Bool",
		SchemaTypeNumber:  "Float64",
		SchemaTypeInteger: "Int64",
		SchemaTypeObject:  "List", // We use List type for objects, because it is compatible with SDKv2
		SchemaTypeArray:   "Set",
		SchemaTypeString:  "String",
	}
}

// typedImport import that depends on type (string
type typedImport string

const (
	planmodifierTypedImport typedImport = "github.com/hashicorp/terraform-plugin-framework/resource/schema/%splanmodifier"
	validatorTypedImport    typedImport = "github.com/hashicorp/terraform-plugin-framework-validators/%svalidator"
	defaultsTypedImport     typedImport = "github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults/%sdefault"
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
