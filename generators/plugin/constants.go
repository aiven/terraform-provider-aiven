package main

import (
	"fmt"
	"strings"
)

// resDatType TF entity type
type resDatType string

const (
	resourceType   resDatType = "resource"
	datasourceType resDatType = "datasource"
)

func boolEntity(isResource bool) resDatType {
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

// Read strings to be formatted with the entity type
type importFmt string

const (
	schemaPackageFmt       importFmt = "github.com/hashicorp/terraform-plugin-framework/%s/schema"
	planmodifierPackageFmt importFmt = "github.com/hashicorp/terraform-plugin-framework/%s/schema/planmodifier"
	timeoutsPackageFmt     importFmt = "github.com/hashicorp/terraform-plugin-framework-timeouts/%s/timeouts"
)

func fmtImport(isResource bool, importString importFmt) string {
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

// getTypedPlanmodifier each plan modifier type has its own package
func getTypedPlanmodifier(kind SchemaType) string {
	suffix := strings.ToLower(typingMapping()[kind])
	return fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework/resource/schema/%splanmodifier", suffix)
}

// getTypedValidator each validator type has its own package
func getTypedValidator(kind SchemaType) string {
	suffix := strings.ToLower(typingMapping()[kind])
	return fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework-validators/%svalidator", suffix)
}
