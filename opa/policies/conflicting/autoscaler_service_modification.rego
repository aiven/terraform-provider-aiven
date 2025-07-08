package aiven.provider.policies.conflicting

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.aiven.provider.policies.utils.terraform as tf

# Check if a resource change is an autoscaler integration deletion
is_autoscaler_deletion(change) if {
	change.type == "aiven_service_integration"
	change.change.actions[_] == "delete"
	change.change.before.integration_type == "autoscaler"
}

# Find services that have autoscaler integrations being removed
services_losing_autoscaler contains service_name if {
	# Find autoscaler integration being destroyed
	some change in tf.all_changes
	is_autoscaler_deletion(change)
	service_name := change.change.before.source_service_name
}

# Find services that are being modified
services_being_modified contains service_name if {
	some change in tf.all_changes
	tf.is_service_update(change)

	# extract service name from before state
	service_name := tf.extract_service_name(change.type, change.change.before)
}

services_being_modified contains service_name if {
	some change in tf.all_changes
	tf.is_service_update(change)

	# extract service name from after state
	service_name := tf.extract_service_name(change.type, change.change.after)
}

# Deny rule: Check for simultaneous autoscaler removal and service modification
#
# METADATA
# title: Aiven Autoscaler Integration and Service Modification Conflict Prevention
# description: Prevents removing autoscaler integrations while simultaneously modifying the associated service in the same Terraform apply
# scope: rule
# schemas:
#   - input: schema.terraform.plan
# related_resources:
#   - https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/service_integration
#   - https://aiven.io/docs/platform/howto/add-storage-space
deny contains msg if {
	# Get services that are losing autoscaler
	service_name := services_losing_autoscaler[_]

	# Check if the same service is also being modified
	service_name in services_being_modified

	msg := sprintf(
		concat("", [
			"POLICY VIOLATION: Autoscaler integration removal and service modification conflict detected. ",
			"Service '%s' has an autoscaler integration being removed while the service itself is being modified in the same apply. ",
			"This can cause inconsistent Terraform plans due to disk space management conflicts. ",
			"Please separate these operations. ",
			"First, remove the autoscaler integration alone. ",
			"For more details please see the docs: https://github.com/aiven/terraform-provider-aiven/blob/main/docs/guides/disk-autoscaler.md",
		]),
		[service_name],
	)
}

# Deny rule for configuration-level detection
deny contains msg if {
	# Find autoscaler integrations in configuration that are not in planned values (being removed)
	integration_config := input.configuration.root_module.resources[_]
	integration_config.type == "aiven_service_integration"

	# get the integration configuration expressions
	config_integration_type := integration_config.expressions.integration_type.constant_value
	config_integration_type == "autoscaler"

	# get the source service name from configuration
	source_service_ref := integration_config.expressions.source_service_name.references[0]

	# Find if this autoscaler integration is being removed (not in planned values)
	integration_being_removed := is_integration_being_removed(integration_config.address)
	integration_being_removed == true

	# check if any service resource references the same service and is being modified
	service_config := input.configuration.root_module.resources[_]
	service_config.type in tf.aiven_service_types

	# check if this service resource is the one referenced by the autoscaler
	service_resource_name := sprintf("%s.%s", [service_config.type, service_config.name])
	service_resource_name == source_service_ref

	# check if this service is being modified
	service_being_modified := is_service_being_modified(service_config.address)
	service_being_modified == true

	msg := sprintf(
		concat("", [
			"POLICY VIOLATION: Autoscaler integration removal and service modification conflict detected. ",
			"Service resource '%s' is being modified while its autoscaler integration '%s' is being removed in the same apply. ",
			"This can cause inconsistent Terraform plans due to disk space management conflicts. ",
			"Please separate these operations. ",
			"For more details please see the docs: https://github.com/aiven/terraform-provider-aiven/blob/main/docs/guides/disk-autoscaler.md",
		]),
		[service_config.address, integration_config.address],
	)
}

is_integration_being_removed(address) := result if {
	some change in tf.all_changes
	change.address == address
	change.change.actions[_] == "delete"
	result := true
} else := false

is_service_being_modified(address) := result if {
	some change in tf.all_changes
	change.address == address
	change.change.actions[_] == "update"
	result := true
} else := false
