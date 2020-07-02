#!/bin/sh

generate-json-schema -f aiven/templates/integration_endpoints_user_config_schema.json > temp/integration_endpoints_user_config.schema.json
generate-json-schema -f aiven/templates/integrations_user_config_schema.json > temp/integrations_user_config.schema.json
generate-json-schema -f aiven/templates/service_user_config_schema.json > temp/service_user_config.schema.json
