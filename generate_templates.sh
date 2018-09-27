#!/usr/bin/bash
avn service types -v --json | jq '[to_entries[] | {"key": .key, "value": .value.user_config_schema}] | from_entries' > templates/service_user_config_schema.json
avn service integration-endpoint-types-list --project test --json | jq 'map({(.endpoint_type): .user_config_schema}) | add' > templates/integration_endpoints_user_config_schema.json
avn service integration-types-list --project test --json | jq 'map({(.integration_type): .user_config_schema}) | add' > templates/integrations_user_config_schema.json
