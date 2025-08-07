package aiven.provider.policies.conflicting

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.aiven.provider.policies.utils.terraform as tf

# all aiven_clickhouse_grant resources from all modules
aiven_clickhouse_grants contains resource if {
	resource := tf.all_resources[_]
	resource.type == "aiven_clickhouse_grant"
}

# collect from resource_changes (for when values are computed)
aiven_clickhouse_grants contains resource if {
	change := tf.all_changes[_]
	change.type == "aiven_clickhouse_grant"
	change.change.after

	# include if values are not available in planned_values
	planned_resource := tf.all_resources[_]
	planned_resource.address == change.address
	planned_resource.values.project == null

	resource := {
		"address": change.address,
		"type": change.type,
		"values": change.change.after,
	}
}

# each role OR user within a service should have only one grant resource
entity_key(resource) := key if {
	# role-based grants
	resource.values.role != null
	key := sprintf("%s:%s:role:%s", [
		resource.values.project,
		resource.values.service_name,
		resource.values.role,
	])
} else := key if {
	# user-based grants
	resource.values.user != null
	key := sprintf("%s:%s:user:%s", [
		resource.values.project,
		resource.values.service_name,
		resource.values.user,
	])
}

# Deny rule: Check for duplicate aiven_clickhouse_grant resources (when values are known)
#
# METADATA
# title: ClickHouse Grant Duplicate Prevention
# description: Prevents duplicate aiven_clickhouse_grant resources for the same role or user in Terraform plans
# scope: rule
# schemas:
#   - input: schema.terraform.plan
# related_resources:
#   - https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/clickhouse_grant
deny contains msg if {
	# get all aiven_clickhouse_grant resources
	resources := [r | some r in aiven_clickhouse_grants]
	count(resources) >= 2

	some i, j
	i < j
	i < count(resources)
	j < count(resources)
	resource1 := resources[i]
	resource2 := resources[j]

	entity_key(resource1) == entity_key(resource2)

	msg := sprintf(
		concat("", [
			"POLICY VIOLATION: Duplicate aiven_clickhouse_grant resources detected. ",
			"Resource '%s' and '%s' both target the same entity '%s'. ",
			"Only one aiven_clickhouse_grant resource is allowed per unique combination of ",
			"project, service_name, and role/user. ",
			"Please consolidate all privileges for this role or user into a single grant resource.",
		]),
		[
			resource1.address,
			resource2.address,
			entity_key(resource1),
		],
	)
}

# configuration-based detection when planned values are not available
deny contains msg if {
	# use config detection if planned_values don't have usable data
	not has_usable_planned_values

	configs := [c |
		c := input.configuration.root_module.resources[_]
		c.type == "aiven_clickhouse_grant"
	]
	count(configs) >= 2

	some i, j
	i < j
	config1 := configs[i]
	config2 := configs[j]

	key1 := config_entity_key(config1)
	key2 := config_entity_key(config2)

	key1 == key2

	msg := sprintf(
		"POLICY VIOLATION: Duplicate aiven_clickhouse_grant resources detected. Resources '%s' and '%s' target the same role or user. Consolidate into one grant resource.",
		[config1.address, config2.address]
	)
}

config_entity_key(config) := key if {
	# role-based grants
	config.expressions.role
	key := sprintf("%v|%v|role:%v", [
		config.expressions.project,
		config.expressions.service_name,
		config.expressions.role
	])
} else := key if {
	# user-based grants
	config.expressions.user
	key := sprintf("%v|%v|user:%v", [
		config.expressions.project,
		config.expressions.service_name,
		config.expressions.user
	])
}

has_usable_planned_values if {
	resource := tf.all_resources[_]
	resource.type == "aiven_clickhouse_grant"
	resource.values.project != null
}
