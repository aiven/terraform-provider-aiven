package aiven.provider.policies.utils.terraform

import future.keywords.contains
import future.keywords.if

# Collect all resources from root module and all child modules recursively
all_resources contains resource if {
	resource := input.planned_values.root_module.resources[_]
}

all_resources contains resource if {
	walk(input.planned_values.root_module, [path, value])
	path[_] == "resources"
	resource := value[_]
}

# Collect all resource changes from root module and all child modules recursively
all_changes contains change if {
	change := input.resource_changes[_]
}

# List of Aiven service resource types
aiven_service_types := {
	"aiven_pg",
	"aiven_mysql",
	"aiven_kafka",
	"aiven_opensearch",
	"aiven_clickhouse",
	"aiven_redis",
	"aiven_cassandra",
	"aiven_grafana",
	"aiven_m3db",
	"aiven_m3aggregator",
	"aiven_dragonfly",
	"aiven_valkey",
	"aiven_thanos",
	"aiven_flink",
	"aiven_alloydbomni",
}

# Helper function to extract service name from service resources
extract_service_name(resource_type, resource) := resource.service_name if {
	# Check if this is a known service type
	resource_type in aiven_service_types

	# Check if the resource actually has a service_name field
	resource.service_name
}

# Check if a resource change is a service update
is_service_update(change) if {
	change.type in aiven_service_types
	change.change.actions[_] == "update"
}
