package aiven.provider.policies.conflicting

import future.keywords.contains
import future.keywords.if
import future.keywords.in

# Import shared utilities
import data.aiven.provider.policies.utils.terraform as tf

# Recursively collect all aiven_organization_permission resources from all modules
aiven_permissions contains resource if {
	resource := tf.all_resources[_]
	resource.type == "aiven_organization_permission"
}

# Create entity key for a resource
entity_key(resource) := key if {
	key := sprintf("%s:%s:%s", [
		resource.values.organization_id,
		resource.values.resource_id,
		resource.values.resource_type,
	])
}

# Deny rule: Check for duplicate aiven_organization_permission resources (when values are known)
#
# METADATA
# title: Aiven Organization Permission Duplicate Prevention
# description: Prevents duplicate aiven_organization_permission resources in Terraform plans
# scope: rule
# schemas:
#   - input: schema.terraform.plan
# related_resources:
#   - https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_permission
deny contains msg if {
	# Get all aiven_organization_permission resources as array
	resources := [r | some r in aiven_permissions]
	count(resources) >= 2

	# Compare each pair of resources
	some i, j
	i < j
	i < count(resources)
	j < count(resources)
	resource1 := resources[i]
	resource2 := resources[j]

	# Check if they target the same entity
	entity_key(resource1) == entity_key(resource2)

	msg := sprintf(
		concat("", [
			"POLICY VIOLATION: Duplicate aiven_organization_permission resources detected. ",
			"Resource '%s' and '%s' both target the same entity '%s'. ",
			"Only one aiven_organization_permission resource is allowed per unique combination of ",
			"organization_id, resource_id, and resource_type. ",
			"Please consolidate these resources into one.",
		]),
		[
			resource1.address,
			resource2.address,
			entity_key(resource1),
		],
	)
}

# ============================================
# Detection based on configuration (for plan-time)
# ============================================

# Get all organization permission resources from configuration
org_permission_configs contains config if {
	input.configuration.root_module.resources
	config := input.configuration.root_module.resources[_]
	config.type == "aiven_organization_permission"
}

# Extract reference values from expressions
get_reference_value(expr) := value if {
	# Handle direct references like aiven_organization.org.id
	expr.references[0]
	value := expr.references[0]
} else := value if {
	# Handle constant values
	expr.constant_value
	value := expr.constant_value
}

# Create a normalized key from configuration
config_key(config) := key if {
	org_expr := config.expressions.organization_id
	res_expr := config.expressions.resource_id
	type_expr := config.expressions.resource_type

	# Extract the base resource being referenced (e.g., "aiven_organization.org")
	org_ref := get_reference_value(org_expr)
	res_ref := get_reference_value(res_expr)
	res_type := get_reference_value(type_expr)

	key := sprintf("%s|%s|%s", [org_ref, res_ref, res_type])
}

# Deny rule for configuration duplicates (plan-time detection)
#
# METADATA
# title: Aiven Organization Permission Duplicate Prevention
# description: Prevents duplicate aiven_organization_permission resources in Terraform plans
# scope: rule
# schemas:
#   - input: schema.terraform.plan
# related_resources:
#   - https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_permission
deny contains msg if {
	configs := [c | c := org_permission_configs[_]]
	count(configs) >= 2

	# Check each pair
	some i, j
	i < j
	config1 := configs[i]
	config2 := configs[j]

	# Compare their keys
	config_key(config1) == config_key(config2)

	msg := sprintf(
		concat("", [
			"POLICY VIOLATION: Duplicate aiven_organization_permission resources detected. ",
			"Resources '%s' and '%s' both target the same entity. ",
			"Only one permission resource is allowed per entity. ",
			"Please consolidate these resources into one.",
		]),
		[
			config1.address,
			config2.address,
		],
	)
}
