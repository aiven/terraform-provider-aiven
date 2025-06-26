package aiven.provider.policies.conflicting

import rego.v1

# Test case 1: Should deny when autoscaler integration is removed and service is modified
test_deny_autoscaler_removal_with_service_modification if {
	deny_messages := deny with input as {
		"resource_changes": [
			{
				"address": "aiven_service_integration.autoscaler_integration",
				"type": "aiven_service_integration",
				"change": {
					"actions": ["delete"],
					"before": {
						"integration_type": "autoscaler",
						"source_service_name": "demopg",
					},
				},
			},
			{
				"address": "aiven_pg.foo",
				"type": "aiven_pg",
				"change": {
					"actions": ["update"],
					"before": {
						"service_name": "demopg",
						"maintenance_window_time": "10:00:00",
					},
					"after": {
						"service_name": "demopg",
						"maintenance_window_time": "11:00:00",
					},
				},
			},
		],
		"planned_values": {"root_module": {"resources": []}},
	}

	count(deny_messages) > 0
	contains(deny_messages[_], "POLICY VIOLATION: Autoscaler integration removal and service modification conflict detected")
	contains(deny_messages[_], "demopg")
}

# Test case 2: Should allow when only autoscaler integration is removed (no service modification)
test_allow_autoscaler_removal_only if {
	deny_messages := deny with input as {
		"resource_changes": [{
			"address": "aiven_service_integration.autoscaler_integration",
			"type": "aiven_service_integration",
			"change": {
				"actions": ["delete"],
				"before": {
					"integration_type": "autoscaler",
					"source_service_name": "demopg",
				},
			},
		}],
		"planned_values": {"root_module": {"resources": []}},
	}

	count(deny_messages) == 0
}

# Test case 3: Should allow when only service is modified (no autoscaler removal)
test_allow_service_modification_only if {
	deny_messages := deny with input as {
		"resource_changes": [{
			"address": "aiven_pg.foo",
			"type": "aiven_pg",
			"change": {
				"actions": ["update"],
				"before": {
					"service_name": "demopg",
					"maintenance_window_time": "10:00:00",
				},
				"after": {
					"service_name": "demopg",
					"maintenance_window_time": "11:00:00",
				},
			},
		}],
		"planned_values": {"root_module": {"resources": []}},
	}

	count(deny_messages) == 0
}

# Test case 4: Should allow when service is created (not updated) and autoscaler is removed
test_allow_service_creation_with_autoscaler_removal if {
	deny_messages := deny with input as {
		"resource_changes": [
			{
				"address": "aiven_service_integration.autoscaler_integration",
				"type": "aiven_service_integration",
				"change": {
					"actions": ["delete"],
					"before": {
						"integration_type": "autoscaler",
						"source_service_name": "demopg",
					},
				},
			},
			{
				"address": "aiven_pg.foo",
				"type": "aiven_pg",
				"change": {
					"actions": ["create"],
					"after": {"service_name": "demopg"},
				},
			},
		],
		"planned_values": {"root_module": {"resources": []}},
	}

	count(deny_messages) == 0
}

# Test case 5: Should deny for different service types
test_deny_mysql_autoscaler_removal_with_service_modification if {
	deny_messages := deny with input as {
		"resource_changes": [
			{
				"address": "aiven_service_integration.mysql_autoscaler",
				"type": "aiven_service_integration",
				"change": {
					"actions": ["delete"],
					"before": {
						"integration_type": "autoscaler",
						"source_service_name": "mysql-service",
					},
				},
			},
			{
				"address": "aiven_mysql.db",
				"type": "aiven_mysql",
				"change": {
					"actions": ["update"],
					"before": {
						"service_name": "mysql-service",
						"plan": "business-8",
					},
					"after": {
						"service_name": "mysql-service",
						"plan": "business-16",
					},
				},
			},
		],
		"planned_values": {"root_module": {"resources": []}},
	}

	count(deny_messages) > 0
	contains(deny_messages[_], "mysql-service")
}

# Test case 6: Should allow non-autoscaler integration removal with service modification
test_allow_non_autoscaler_integration_removal if {
	deny_messages := deny with input as {
		"resource_changes": [
			{
				"address": "aiven_service_integration.datadog_integration",
				"type": "aiven_service_integration",
				"change": {
					"actions": ["delete"],
					"before": {
						"integration_type": "datadog",
						"source_service_name": "demopg",
					},
				},
			},
			{
				"address": "aiven_pg.foo",
				"type": "aiven_pg",
				"change": {
					"actions": ["update"],
					"before": {
						"service_name": "demopg",
						"maintenance_window_time": "10:00:00",
					},
					"after": {
						"service_name": "demopg",
						"maintenance_window_time": "11:00:00",
					},
				},
			},
		],
		"planned_values": {"root_module": {"resources": []}},
	}

	count(deny_messages) == 0
}
