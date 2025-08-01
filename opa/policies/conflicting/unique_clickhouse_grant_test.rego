package aiven.provider.policies.conflicting

import future.keywords.contains
import future.keywords.if

# Test: No duplicates - single resource
test_no_duplicates_single_resource if {
	test_input := plan_with_resources([clickhouse_grant("aiven_clickhouse_grant.single", "test-project", "test-service", "dba")])

	count(deny) == 0 with input as test_input
}

# Test: No duplicates - different roles
test_no_duplicates_different_roles if {
	test_input := plan_with_resources([
		clickhouse_grant("aiven_clickhouse_grant.dba", "test-project", "test-service", "dba"),
		clickhouse_grant("aiven_clickhouse_grant.readonly", "test-project", "test-service", "readonly"),
	])

	count(deny) == 0 with input as test_input
}

# Test: No duplicates - different services
test_no_duplicates_different_services if {
	test_input := plan_with_resources([
		clickhouse_grant("aiven_clickhouse_grant.service1_dba", "test-project", "service-1", "dba"),
		clickhouse_grant("aiven_clickhouse_grant.service2_dba", "test-project", "service-2", "dba"),
	])

	count(deny) == 0 with input as test_input
}

# Test: No duplicates - different projects
test_no_duplicates_different_projects if {
	test_input := plan_with_resources([
		clickhouse_grant("aiven_clickhouse_grant.project1_dba", "project-1", "test-service", "dba"),
		clickhouse_grant("aiven_clickhouse_grant.project2_dba", "project-2", "test-service", "dba"),
	])

	count(deny) == 0 with input as test_input
}

# Test: Duplicate grants for same role
test_duplicate_grants_same_role if {
	test_input := plan_with_resources([
		clickhouse_grant("aiven_clickhouse_grant.dba_grants_1", "test-project", "test-service", "dba"),
		clickhouse_grant("aiven_clickhouse_grant.dba_grants_2", "test-project", "test-service", "dba"),
	])

	count(deny) == 1 with input as test_input
	contains(deny[_], "POLICY VIOLATION: Duplicate aiven_clickhouse_grant resources detected") with input as test_input
	contains(deny[_], "aiven_clickhouse_grant.dba_grants_1") with input as test_input
	contains(deny[_], "aiven_clickhouse_grant.dba_grants_2") with input as test_input
}

# Test: Multiple duplicates
test_multiple_duplicates if {
	test_input := plan_with_resources([
		clickhouse_grant("aiven_clickhouse_grant.dba_grants_1", "test-project", "test-service", "dba"),
		clickhouse_grant("aiven_clickhouse_grant.dba_grants_2", "test-project", "test-service", "dba"),
		clickhouse_grant("aiven_clickhouse_grant.dba_grants_3", "test-project", "test-service", "dba"),
	])

	count(deny) == 3 with input as test_input # should detect 3 pairs
}

# Test: Mixed resources - only ClickHouse grants checked
test_mixed_resources if {
	test_input := plan_with_resources([
		clickhouse_grant("aiven_clickhouse_grant.dba_grants_1", "test-project", "test-service", "dba"),
		clickhouse_grant("aiven_clickhouse_grant.dba_grants_2", "test-project", "test-service", "dba"),
		some_resource("aiven_clickhouse.service", "aiven_clickhouse"),
		some_resource("aiven_clickhouse_role.dba", "aiven_clickhouse_role"),
	])

	count(deny) == 1 with input as test_input
	contains(deny[_], "aiven_clickhouse_grant.dba_grants_1") with input as test_input
	contains(deny[_], "aiven_clickhouse_grant.dba_grants_2") with input as test_input
}

# Test: Configuration-level duplicate detection
test_config_duplicate_grants if {
	test_input := config_with_resources([
		clickhouse_grant_config(
			"aiven_clickhouse_grant.dba_grants_1",
			{"constant_value": "test-project"},
			{"references": ["aiven_clickhouse.foo.service_name"]},
			{"references": ["aiven_clickhouse_role.dba.role"]},
		),
		clickhouse_grant_config(
			"aiven_clickhouse_grant.dba_grants_2",
			{"constant_value": "test-project"},
			{"references": ["aiven_clickhouse.foo.service_name"]},
			{"references": ["aiven_clickhouse_role.dba.role"]},
		),
	])

	count(deny) == 1 with input as test_input
	contains(deny[_], "POLICY VIOLATION: Duplicate aiven_clickhouse_grant resources detected") with input as test_input
	contains(deny[_], "aiven_clickhouse_grant.dba_grants_1") with input as test_input
	contains(deny[_], "aiven_clickhouse_grant.dba_grants_2") with input as test_input
}

# Test: Configuration-level different roles
test_config_different_roles if {
	test_input := config_with_resources([
		clickhouse_grant_config(
			"aiven_clickhouse_grant.dba_grants",
			{"constant_value": "test-project"},
			{"references": ["aiven_clickhouse.foo.service_name"]},
			{"references": ["aiven_clickhouse_role.dba.role"]},
		),
		clickhouse_grant_config(
			"aiven_clickhouse_grant.readonly_grants",
			{"constant_value": "test-project"},
			{"references": ["aiven_clickhouse.foo.service_name"]},
			{"references": ["aiven_clickhouse_role.readonly.role"]},
		),
	])

	count(deny) == 0 with input as test_input
}

# Test: Configuration-level different services
test_config_different_services if {
	test_input := config_with_resources([
		clickhouse_grant_config(
			"aiven_clickhouse_grant.service1_dba",
			{"constant_value": "test-project"},
			{"references": ["aiven_clickhouse.service1.service_name"]},
			{"references": ["aiven_clickhouse_role.dba.role"]},
		),
		clickhouse_grant_config(
			"aiven_clickhouse_grant.service2_dba",
			{"constant_value": "test-project"},
			{"references": ["aiven_clickhouse.service2.service_name"]},
			{"references": ["aiven_clickhouse_role.dba.role"]},
		),
	])

	count(deny) == 0 with input as test_input
}

# Test: User grants - no duplicates
test_no_duplicates_user_grants if {
	test_input := plan_with_resources([
		clickhouse_user_grant("aiven_clickhouse_grant.user1_grants", "test-project", "test-service", "user1"),
		clickhouse_user_grant("aiven_clickhouse_grant.user2_grants", "test-project", "test-service", "user2"),
	])

	count(deny) == 0 with input as test_input
}

# Test: Duplicate user grants
test_duplicate_user_grants if {
	test_input := plan_with_resources([
		clickhouse_user_grant("aiven_clickhouse_grant.user_grants_1", "test-project", "test-service", "analyst"),
		clickhouse_user_grant("aiven_clickhouse_grant.user_grants_2", "test-project", "test-service", "analyst"),
	])

	count(deny) == 1 with input as test_input
	contains(deny[_], "POLICY VIOLATION: Duplicate aiven_clickhouse_grant resources detected") with input as test_input
	contains(deny[_], "aiven_clickhouse_grant.user_grants_1") with input as test_input
	contains(deny[_], "aiven_clickhouse_grant.user_grants_2") with input as test_input
}

# Test: Mixed role and user grants
test_mixed_role_user_grants if {
	test_input := plan_with_resources([
		clickhouse_grant("aiven_clickhouse_grant.role_grants", "test-project", "test-service", "dba"),
		clickhouse_user_grant("aiven_clickhouse_grant.user_grants", "test-project", "test-service", "dba"),
	])

	count(deny) == 0 with input as test_input
}

# create aiven_clickhouse_grant resource for roles
clickhouse_grant(address, project, service_name, role) := {
	"address": address,
	"type": "aiven_clickhouse_grant",
	"values": {
		"project": project,
		"service_name": service_name,
		"role": role,
		"user": null,
	},
}

# create aiven_clickhouse_grant resource for users
clickhouse_user_grant(address, project, service_name, user) := {
	"address": address,
	"type": "aiven_clickhouse_grant",
	"values": {
		"project": project,
		"service_name": service_name,
		"role": null,
		"user": user,
	},
}

# create other resource types
some_resource(address, res_type) := {
	"address": address,
	"type": res_type,
	"values": {
		"id": "some-id",
		"name": "some-name",
	},
}

# create plan input
plan_with_resources(resources) := {"planned_values": {"root_module": {"resources": resources}}}

# create config input
config_with_resources(resources) := {"configuration": {"root_module": {"resources": resources}}}

# create config resource with expressions for roles
clickhouse_grant_config(address, project_expr, service_expr, role_expr) := {
	"address": address,
	"type": "aiven_clickhouse_grant",
	"expressions": {
		"project": project_expr,
		"service_name": service_expr,
		"role": role_expr,
	},
}

# create config resource with expressions for users
clickhouse_user_grant_config(address, project_expr, service_expr, user_expr) := {
	"address": address,
	"type": "aiven_clickhouse_grant",
	"expressions": {
		"project": project_expr,
		"service_name": service_expr,
		"user": user_expr,
	},
}
