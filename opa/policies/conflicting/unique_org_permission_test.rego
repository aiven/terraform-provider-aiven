package aiven.provider.policies.conflicting

import future.keywords.contains
import future.keywords.if

# Test helper to create aiven_organization_permission resource
org_permission(address, org_id, res_id, res_type) := {
	"address": address,
	"type": "aiven_organization_permission",
	"values": {
		"organization_id": org_id,
		"resource_id": res_id,
		"resource_type": res_type,
	},
}

# Test helper to create other resource types
some_resource(address, res_type) := {
	"address": address,
	"type": res_type,
	"values": {
		"id": "some-id",
		"name": "some-name",
	},
}

# Test helper to create plan input
plan_with_resources(resources) := {"planned_values": {"root_module": {"resources": resources}}}

# Test helper to create plan with nested modules
plan_with_modules(root_resources, child_resources) := {"planned_values": {"root_module": {
	"resources": root_resources,
	"child_modules": [{"resources": child_resources}],
}}}

# Test: No duplicates - single resource
test_no_duplicates_single_resource if {
	test_input := plan_with_resources([org_permission("aiven_organization_permission.single", "org1", "res1", "project")])

	count(deny) == 0 with input as test_input
}

# Test: No duplicates - different organizations
test_no_duplicates_different_orgs if {
	test_input := plan_with_resources([
		org_permission("aiven_organization_permission.perm1", "org1", "res1", "project"),
		org_permission("aiven_organization_permission.perm2", "org2", "res1", "project"),
	])

	count(deny) == 0 with input as test_input
}

# Test: No duplicates - different resources
test_no_duplicates_different_resources if {
	test_input := plan_with_resources([
		org_permission("aiven_organization_permission.perm1", "org1", "res1", "project"),
		org_permission("aiven_organization_permission.perm2", "org1", "res2", "project"),
	])

	count(deny) == 0 with input as test_input
}

# Test: No duplicates - different resource types
test_no_duplicates_different_types if {
	test_input := plan_with_resources([
		org_permission("aiven_organization_permission.perm1", "org1", "res1", "project"),
		org_permission("aiven_organization_permission.perm2", "org1", "res1", "organization"),
	])

	count(deny) == 0 with input as test_input
}

# Test: Duplicates - exact match
test_duplicates_exact_match if {
	test_input := plan_with_resources([
		org_permission("aiven_organization_permission.perm1", "org1", "res1", "project"),
		org_permission("aiven_organization_permission.perm2", "org1", "res1", "project"),
	])

	violations := deny with input as test_input
	count(violations) == 1
	violation := violations[_]
	contains(violation, "POLICY VIOLATION")
	contains(violation, "perm1")
	contains(violation, "perm2")
}

# Test: Multiple duplicate pairs
test_multiple_duplicate_pairs if {
	test_input := plan_with_resources([
		org_permission("aiven_organization_permission.perm1", "org1", "res1", "project"),
		org_permission("aiven_organization_permission.perm2", "org1", "res1", "project"),
		org_permission("aiven_organization_permission.perm3", "org2", "res2", "organization"),
		org_permission("aiven_organization_permission.perm4", "org2", "res2", "organization"),
	])

	violations := deny with input as test_input
	count(violations) == 2
}

# Test: Duplicates with mixed resource types
test_duplicates_with_mixed_resources if {
	test_input := plan_with_resources([
		some_resource("aiven_organization.org", "aiven_organization"),
		org_permission("aiven_organization_permission.perm1", "org1", "res1", "project"),
		some_resource("aiven_organization_user_group.group", "aiven_organization_user_group"),
		org_permission("aiven_organization_permission.perm2", "org1", "res1", "project"),
	])

	violations := deny with input as test_input
	count(violations) == 1
	violation := violations[_]
	contains(violation, "POLICY VIOLATION")
}

# Test: No permission resources
test_no_permission_resources if {
	test_input := plan_with_resources([
		some_resource("aiven_organization.org", "aiven_organization"),
		some_resource("aiven_project.project", "aiven_project"),
	])

	count(deny) == 0 with input as test_input
}

# Test: Duplicates across nested modules
test_nested_modules_duplicates if {
	test_input := plan_with_modules(
		[org_permission("aiven_organization_permission.root_perm", "org1", "res1", "project")],
		[org_permission("module.child.aiven_organization_permission.child_perm", "org1", "res1", "project")],
	)

	violations := deny with input as test_input
	count(violations) == 1
	violation := violations[_]
	contains(violation, "POLICY VIOLATION")
}

# Test: Triple duplicates (should generate 3 violations - one for each pair)
test_triple_duplicates if {
	test_input := plan_with_resources([
		org_permission("aiven_organization_permission.perm1", "org1", "res1", "project"),
		org_permission("aiven_organization_permission.perm2", "org1", "res1", "project"),
		org_permission("aiven_organization_permission.perm3", "org1", "res1", "project"),
	])

	violations := deny with input as test_input
	count(violations) == 3 # C(3,2) = 3 pairs
}

# Test: Complex scenario with real-world data
test_complex_real_world_scenario if {
	test_input := {"planned_values": {"root_module": {"resources": [
		{
			"address": "aiven_organization.org",
			"mode": "managed",
			"type": "aiven_organization",
			"name": "org",
			"provider_name": "registry.terraform.io/aiven/aiven",
			"schema_version": 0,
			"values": {
				"create_time": "2025-06-11T08:03:41Z",
				"id": "orgXXXXXXXXXX",
				"name": "test-org-1",
				"tenant_id": "aiven",
				"timeouts": null,
				"update_time": "2025-06-11T08:03:41Z",
			},
			"sensitive_values": {},
		},
		{
			"address": "aiven_organization_permission.duplicate",
			"mode": "managed",
			"type": "aiven_organization_permission",
			"name": "duplicate",
			"provider_name": "registry.terraform.io/aiven/aiven",
			"schema_version": 0,
			"values": {
				"organization_id": "orgXXXXXXXXXX",
				"permissions": [{
					"permissions": ["organization:domains:write"],
					"principal_id": "ugXXXXXXXXXX",
					"principal_type": "user_group",
				}],
				"resource_id": "orgXXXXXXXXXX",
				"resource_type": "organization",
				"timeouts": null,
			},
			"sensitive_values": {},
		},
		{
			"address": "aiven_organization_permission.example_org_permissions",
			"mode": "managed",
			"type": "aiven_organization_permission",
			"name": "example_org_permissions",
			"provider_name": "registry.terraform.io/aiven/aiven",
			"schema_version": 0,
			"values": {
				"id": "orgXXXXXXXXXX/organization/orgXXXXXXXXXX",
				"organization_id": "orgXXXXXXXXXX",
				"permissions": [{
					"create_time": "2025-06-11T08:03:42Z",
					"permissions": ["organization:users:write"],
					"principal_id": "ugXXXXXXXXXX",
					"principal_type": "user_group",
					"update_time": "2025-06-11T08:05:43Z",
				}],
				"resource_id": "orgXXXXXXXXXX",
				"resource_type": "organization",
				"timeouts": null,
			},
			"sensitive_values": {},
		},
		{
			"address": "aiven_organization_user_group.example",
			"mode": "managed",
			"type": "aiven_organization_user_group",
			"name": "example",
			"provider_name": "registry.terraform.io/aiven/aiven",
			"schema_version": 0,
			"values": {
				"create_time": "2025-06-11T08:03:42Z",
				"description": "Example group of users.",
				"group_id": "ugXXXXXXXXXX",
				"id": "orgXXXXXXXXXX/ugXXXXXXXXXX",
				"name": "Example group",
				"organization_id": "orgXXXXXXXXXX",
				"timeouts": null,
				"update_time": "2025-06-11T08:03:42Z",
			},
			"sensitive_values": {},
		},
	]}}}

	violations := deny with input as test_input
	count(violations) == 1
	violation := violations[_]
	contains(violation, "POLICY VIOLATION")
	contains(violation, "duplicate")
	contains(violation, "example_org_permissions")
}

# Test to verify the policy correctly identifies the duplicate pattern with unknown referenced values
test_configuration_based_detection if {
	test_input := {
		"configuration": {"root_module": {"resources": [
			{
				"address": "aiven_organization_permission.perm1",
				"type": "aiven_organization_permission",
				"expressions": {
					"organization_id": {"references": ["aiven_project.foo.id"]},
					"resource_id": {"references": ["aiven_project.foo.id"]},
					"resource_type": {"constant_value": "project"},
				},
			},
			{
				"address": "aiven_organization_permission.perm2",
				"type": "aiven_organization_permission",
				"expressions": {
					"organization_id": {"references": ["aiven_project.foo.id"]},
					"resource_id": {"references": ["aiven_project.foo.id"]},
					"resource_type": {"constant_value": "project"},
				},
			},
		]}},
		"planned_values": {"root_module": {"resources": []}},
	}

	violations := deny with input as test_input
	count(violations) == 1
}

# Test: Empty resources array
test_empty_resources if {
	test_input := plan_with_resources([])
	count(deny) == 0 with input as test_input
}
