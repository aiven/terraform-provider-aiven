---
title: Changelog
parent: README
nav_order: 1
---

# Changelog

<!-- Always keep the following header in place: -->
<!--## [MAJOR.MINOR.PATCH] - YYYY-MM-DD -->

## [4.37.0] - 2025-03-12

- Fix user config plan suppresses `false` values 
- Fix `terraform import` for various resources
- Fix `aiven_organization_application_user`: improve delete operation
- Improve `aiven_organiztion_permission`: use `PermissionsSet` endpoint to set permissions
- Restricted `organization_group_project` resource modification
- Restricted `project_user` resource modification
- Add `aiven_organization_application_user_token` resource field `ip_allowlist`: List of allowed IP ranges.
- Change `aiven_account_team_project` field `team_type` (enum): add `organization:billing:read`, `organization:billing:write`
- Add `aiven_kafka` field `kafka_user_config.kafka_rest_config.consumer_idle_disconnect_timeout`: Specifies the maximum
  duration (in seconds) a client can remain idle before it is deleted
- Change `aiven_organization_permission` resource field `permissions.permissions` (enum): add `organization:billing:read`, `organization:billing:write`
- Change `aiven_project_user` field `member_type` (enum): add `organization:billing:read`, `organization:billing:write`
- Change `aiven_service_integration` field `clickhouse_kafka_user_config.tables`: max items ~~`100`~~ → `400`
- Add `aiven_valkey` field `valkey_user_config.enable_ipv6`: Register AAAA DNS records for the service, and allow IPv6
  packets to service ports.
- Add `aiven_valkey` field `valkey_user_config.valkey_active_expire_effort`: Valkey reclaims expired keys both when
  accessed and in the background
- Add `aiven_opensearch` field `opensearch_user_config.opensearch.disk_watermarks`: Watermark settings
- Add `aiven_opensearch` field `opensearch_user_config.azure_migration.readonly`: Whether the repository is read-only.
  Default: `false`.
- Add `aiven_opensearch` field `opensearch_user_config.gcs_migration.readonly`: Whether the repository is read-only.
  Default: `false`.
- Add `aiven_opensearch` field `opensearch_user_config.s3_migration.readonly`: Whether the repository is read-only.
  Default: `false`.

## [4.36.0] - 2025-02-20

- Added `aiven_organization_vpc` beta resource and data source
- Added `aiven_aws_org_vpc_peering_connection` beta resource and data source
- Added `aiven_gcp_org_vpc_peering_connection` beta resource and data source
- Added `aiven_azure_org_vpc_peering_connection` beta resource and data source

## [4.35.0] - 2025-02-18

- Fix `aiven_service_integration` type `clickhouse_kafka` crashes when a table is removed
- Added `aiven_organization_project` beta resource

## [4.34.2] - 2025-02-13

- Fix `ip_filter`, suppress plan for the new `ip_filter` default value (`::/0`) when a new service is created
- Change `aiven_opensearch` field `opensearch_user_config.opensearch.auth_failure_listeners.ip_rate_limiting`: deprecate:
  This property is deprecated

## [4.34.1] - 2025-02-07

- Fix `organization_user_group` import

## [4.34.0] - 2025-01-29

- Fix `pg_user`: throws an error when the password is modified directly in the service
- Enable `read_replica` service integration for `aiven_redis` and `aiven_valkey` resources
- Add `disaster_recovery` service integration type support
- Add `aiven_flink_jar_application`, `aiven_flink_jar_application_version` and `aiven_flink_jar_application_deployment` BETA resources
- Fix: services always created in the default VPC of the same cloud, when `project_vpc_id` wasn't set
- Change `aiven_account_team_project` field `team_type` (enum): add `organization:networking:read`, `organization:networking:write`
- Change `aiven_organization_permission` resource field `permissions.permissions` (enum):
  add `organization:networking:read`, `organization:networking:write`
- Change `aiven_project_user` field `member_type` (enum): add `organization:networking:read`, `organization:networking:write`
- Change `aiven_service_integration_endpoint` field `endpoint_type` (enum): add `external_azure_blob_storage`
- Replaced `aiven-go-client/v2` with `aiven/go-client-codegen` in `project_user` resource/data source
- Add `aiven_valkey` field `valkey_user_config.frequent_snapshots`: When enabled, Valkey will create frequent local RDB snapshots

## [4.33.0] - 2025-01-16

- Fix `pg_user` gets 404 error on create
- Add `aiven_flink` resource and datasource field `flink_user_config.custom_code`: Enable to upload Custom JARs for Flink applications.
- Add `aiven_grafana` resource and datasource field `grafana_user_config.dashboard_scenes_enabled`: Enable
  use of the Grafana Scenes Library as the dashboard engine
- Add `aiven_kafka_connect` resource and datasource field `kafka_connect_user_config.plugin_versions`: The plugin selected by the user
- Add `aiven_opensearch` resource and datasource field `opensearch_user_config.opensearch.cluster_search_request_slowlog`
- Add `aiven_opensearch` resource and datasource field `opensearch_user_config.opensearch.enable_remote_backed_storage`:
  Enable remote-backed storage.
- Add `aiven_service_integration_endpoint` resource and datasource field `external_azure_blob_storage_user_config`:
  ExternalAzureBlobStorage user configurable settings

## [4.32.0] - 2025-01-14

- Add `aiven_kafka_quota` resource
- Add `aiven_opensearch` resource and datasource field
  `opensearch_user_config.opensearch.cluster_routing_allocation_balance_prefer_primary`: When set to true, OpenSearch
  attempts to evenly distribute the primary shards between the cluster nodes
- Add `aiven_opensearch` resource and datasource field `opensearch_user_config.opensearch.segrep`: Segment Replication
  Backpressure Settings
- Replaced `aiven-go-client/v2` with `aiven/go-client-codegen` in `account` resource/data source
- Replaced `aiven-go-client/v2` with `aiven/go-client-codegen` in `account_authentication` resource/data source
- Replaced `aiven-go-client/v2` with `aiven/go-client-codegen` in `account_team` resource/data source
- Replaced `aiven-go-client/v2` with `aiven/go-client-codegen` in `account_team_member` resource/data source
- Replaced `aiven-go-client/v2` with `aiven/go-client-codegen` in `account_team_project` resource/data source
- Replaced `aiven-go-client/v2` with `aiven/go-client-codegen` in `aiven_pg_user` resource/data source
- Fix `aiven_pg_user` creating in bulk occasionally results in a 404 error
- Use `aiven_pg_user` handlers and schema in `aiven_alloydbomni_user` resource
- Ensure the service user's password is updated correctly

## [4.31.1] - 2024-12-23

- Validate whether the `aiven_project.billing_group` field has changed before calling admin API

## [4.31.0] - 2024-12-18

- Add `alloydbomni` BETA resource and datasource
- Add `aiven_alloydbomni_user` BETA resource and datasource
- Add `aiven_alloydbomni_database` BETA resource and datasource
- Fix `terraform plan`: new resources don't display zero values for user configuration options
- Add `aiven_service_integration` resource and datasource field `destination_service_project`: Destination service project name
- Add `aiven_service_integration` resource and datasource field `source_service_project`: Source service project name
- Change `aiven_account_team_project` resource and datasource field `team_type` (enum): remove
  `organization:billing:read`, `organization:billing:write`, `organization:network:read`, `organization:network:write`,
  `organization:permissions:read`, `organization:permissions:write`, `organization:projects:read`, `organization:projects:write`
- Add `aiven_alloydbomni` resource and datasource field `alloydbomni_user_config.pg.password_encryption` (enum)
- Change `aiven_flink` resource and datasource field `flink_user_config.flink_version` (enum): add `1.20`
- Add `aiven_opensearch` resource and datasource field
  `opensearch_user_config.opensearch_dashboards.multiple_data_source_enabled`: Enable or disable multiple data sources
  in OpenSearch Dashboards
- Change `aiven_organization_permission` resource field `permissions.permissions` (enum): remove
  `organization:billing:read`, `organization:billing:write`, `organization:network:read`, `organization:network:write`,
  `organization:permissions:read`, `organization:permissions:write`, `organization:projects:read`, `organization:projects:write`
- Add `aiven_pg` resource and datasource field `pg_user_config.pg.password_encryption` (enum)
- Change `aiven_pg` resource and datasource field `pg_user_config.additional_backup_regions`: remove deprecation
- Change `aiven_pg` resource and datasource field `pg_user_config.pg_version` (enum): add `17`
- Change `aiven_project_user` resource and datasource field `member_type` (enum): remove `organization:billing:read`,
  `organization:billing:write`, `organization:network:read`, `organization:network:write`,
  `organization:permissions:read`, `organization:permissions:write`, `organization:projects:read`, `organization:projects:write`

## [4.30.0] - 2024-12-05

- Add `aiven_kafka_native_acl` resource
- Add `aiven_opensearch` resource field `opensearch_user_config.opensearch.search_insights_top_queries`
- Add `aiven_thanos` resource field `thanos_user_config.private_access`: Allow access to selected service ports from
  private networks
- Add `aiven_thanos` resource field `thanos_user_config.privatelink_access`: Allow access to selected service components
  through Privatelink
- Add `aiven_opensearch` datasource field `opensearch_user_config.opensearch.search_insights_top_queries`
- Add `aiven_thanos` datasource field `thanos_user_config.private_access`: Allow access to selected service ports from
  private networks
- Add `aiven_thanos` datasource field `thanos_user_config.privatelink_access`: Allow access to selected service components
  through Privatelink
- Add `aiven_service_integration_endpoint` resource field `datadog_user_config.extra_tags_prefix`: Extra tags prefix.
  Defaults to aiven.
- Add `aiven_service_integration_endpoint` datasource field `datadog_user_config.extra_tags_prefix`: Extra tags prefix.
  Defaults to aiven.
- Add `exactly_once_delivery_enabled` field to `aiven_mirrormaker_replication_flow`: Whether to enable exactly-once message delivery for Mirrormaker replication flow

## [4.29.0] - 2024-11-14

- Add support for `autoscaler` service integration
- Add `aiven_opensearch` resource field `opensearch_user_config.azure_migration.include_aliases`: Whether to restore
  aliases alongside their associated indexes
- Add `aiven_opensearch` resource field `opensearch_user_config.gcs_migration.include_aliases`: Whether to restore aliases
  alongside their associated indexes
- Add `aiven_opensearch` resource field `opensearch_user_config.s3_migration.include_aliases`: Whether to restore aliases
  alongside their associated indexes
- Add `aiven_opensearch` datasource field `opensearch_user_config.gcs_migration.include_aliases`: Whether to restore
  aliases alongside their associated indexes
- Add `aiven_opensearch` datasource field `opensearch_user_config.azure_migration.include_aliases`: Whether to restore
  aliases alongside their associated indexes
- Add `aiven_opensearch` datasource field `opensearch_user_config.s3_migration.include_aliases`: Whether to restore
  aliases alongside their associated indexes
- Change `aiven_cassandra` resource field `cassandra_user_config.additional_backup_regions`: remove deprecation
- Change `aiven_cassandra` datasource field `cassandra_user_config.additional_backup_regions`: remove deprecation
- Add `aiven_influxdb` resource field `influxdb_user_config.private_access.user_backup`: Allow clients to connect
  to user_backup with a DNS name that always resolves to the service's private IP addresses
- Add `aiven_influxdb` resource field `influxdb_user_config.privatelink_access.user_backup`: Enable user_backup.
- Add `aiven_influxdb` resource field `influxdb_user_config.public_access.user_backup`: Allow clients to connect
  to user_backup from the public internet for service nodes that are in a project VPC or another type of private network.
- Add `aiven_thanos` resource field `thanos_user_config.receiver_ingesting`: CommonReceive.
- Add `aiven_thanos` resource field `thanos_user_config.receiver_routing`: ThanosReceiveRouting.
- Add `aiven_thanos` resource field `thanos_user_config.ruler`: ThanosRuler.
- Add `aiven_thanos` resource field `thanos_user_config.store`: ThanosStore.
- Add `aiven_influxdb` datasource field `influxdb_user_config.private_access.user_backup`: Allow clients to connect
  to user_backup with a DNS name that always resolves to the service's private IP addresses
- Add `aiven_influxdb` datasource field `influxdb_user_config.privatelink_access.user_backup`: Enable user_backup.
- Add `aiven_influxdb` datasource field `influxdb_user_config.public_access.user_backup`: Allow clients to connect
  to user_backup from the public internet for service nodes that are in a project VPC or another type of private network.
- Add `aiven_thanos` datasource field `thanos_user_config.receiver_ingesting`: CommonReceive.
- Add `aiven_thanos` datasource field `thanos_user_config.receiver_routing`: ThanosReceiveRouting.
- Add `aiven_thanos` datasource field `thanos_user_config.ruler`: ThanosRuler.
- Add `aiven_thanos` datasource field `thanos_user_config.store`: ThanosStore.
- Change `aiven_account_team_project` resource field `team_type`: add `organization:app_users:write`,
  `organization:audit_logs:read`, `organization:billing:read`, `organization:billing:write`, `organization:domains:write`,
  `organization:groups:write`, `organization:idps:write`, `organization:network:read`, `organization:network:write`,
  `organization:permissions:read`, `organization:permissions:write`, `organization:projects:read`,
  `organization:projects:write`, `organization:users:write`, `project:services:write`, `role:organization:admin`,
  `role:services:maintenance`, `role:services:recover`, `service:data:write`, `service:secrets:read`,
  `service:users:write`, remove `services:maintenance`
- Change `aiven_opensearch` resource field `opensearch_user_config.azure_migration.indices`: optional ~~`true`~~ → `false`
- Change `aiven_opensearch` resource field `opensearch_user_config.gcs_migration.indices`: optional ~~`true`~~ → `false`
- Change `aiven_opensearch` resource field `opensearch_user_config.s3_migration.indices`: optional ~~`true`~~ → `false`
- Change `aiven_organization_permission` resource field `permissions.permissions`: add `organization:app_users:write`,
  `organization:audit_logs:read`, `organization:billing:read`, `organization:billing:write`, `organization:domains:write`,
  `organization:groups:write`, `organization:idps:write`, `organization:network:read`, `organization:network:write`,
  `organization:permissions:read`, `organization:permissions:write`, `organization:projects:read`,
  `organization:projects:write`, `organization:users:write`, `project:services:write`, `role:organization:admin`,
  `role:services:maintenance`, `role:services:recover`, `service:data:write`, `service:secrets:read`,
  `service:users:write`, remove `services:maintenance`
- Change `aiven_organization_permission` resource field `resource_type`: add `organization_unit`, `organization`
- Change `aiven_project_user` resource field `member_type`: add `organization:app_users:write`,
  `organization:audit_logs:read`, `organization:billing:read`, `organization:billing:write`, `organization:domains:write`,
  `organization:groups:write`, `organization:idps:write`, `organization:network:read`, `organization:network:write`,
  `organization:permissions:read`, `organization:permissions:write`, `organization:projects:read`,
  `organization:projects:write`, `organization:users:write`, `project:services:write`, `role:organization:admin`,
  `role:services:maintenance`, `role:services:recover`, `service:data:write`, `service:secrets:read`,
  `service:users:write`, remove `services:maintenance`
- Change `aiven_account_team_project` datasource field `team_type`: add `organization:app_users:write`,
  `organization:audit_logs:read`, `organization:billing:read`, `organization:billing:write`, `organization:domains:write`,
  `organization:groups:write`, `organization:idps:write`, `organization:network:read`, `organization:network:write`,
  `organization:permissions:read`, `organization:permissions:write`, `organization:projects:read`,
  `organization:projects:write`, `organization:users:write`, `project:services:write`, `role:organization:admin`,
  `role:services:maintenance`, `role:services:recover`, `service:data:write`, `service:secrets:read`,
  `service:users:write`, remove `services:maintenance`
- Change `aiven_opensearch` datasource field `opensearch_user_config.azure_migration.indices`: optional ~~`true`~~ → `false`
- Change `aiven_opensearch` datasource field `opensearch_user_config.gcs_migration.indices`: optional ~~`true`~~ → `false`
- Change `aiven_opensearch` datasource field `opensearch_user_config.s3_migration.indices`: optional ~~`true`~~ → `false`
- Change `aiven_project_user` datasource field `member_type`: add `organization:app_users:write`,
  `organization:audit_logs:read`, `organization:billing:read`, `organization:billing:write`, `organization:domains:write`,
  `organization:groups:write`, `organization:idps:write`, `organization:network:read`, `organization:network:write`,
  `organization:permissions:read`, `organization:permissions:write`, `organization:projects:read`,
  `organization:projects:write`, `organization:users:write`, `project:services:write`, `role:organization:admin`,
  `role:services:maintenance`, `role:services:recover`, `service:data:write`, `service:secrets:read`,
  `service:users:write`, remove `services:maintenance`

## [4.28.0] - 2024-10-21

- Fix `aiven_project`: can't migrate from `account_id` (deprecated) to `parent_id`
- Add `aiven_organization_user_list` beta resource
- Run client-side validation for `aiven_kafka_schema` AVRO type schema

## [4.27.0] - 2024-10-09

- Remove `aiven_thanos` from beta resources
- Removes `receiver_ingesting_remote_write_uri` and `store_uri` Thanos connection info fields
- Adds `stringtype` to `flink_external_postgresql_user_config` service integration
- Fix `terraform import` for services with additional disk space or read replica service integration

## [4.26.0] - 2024-09-25

- Remove `aiven_valkey` from beta resources
- Remove `aiven_valkey_user` from beta resources
- Adds`aiven_organization_permission` example
- Add capability to map external service user with internal aiven user with external_identity data source

## [4.25.0] - 2024-09-17

- Fix `aiven_kafka_connector`: increase create polling timeout
- Add experimental `aiven_organization_permission` resource
- Add `prefix_path_depth` to `kafka_user_config.kafka_connect_secret_providers.vault`
- Add `backup_hour` and `backup_minute` to `redis_user_config` and `valkey_user_config`
- Add user configuration options for `external_s3`, `external_clickhouse` and `external_mysql` integration endpoint types
- Use `ServiceGet` from the code-generated client
- Use the code-generated client to manage `aiven_service_integration` and `aiven_service_integration_endpoint`
- Use Go 1.23
- Add capability to set description and owner group per `aiven_kafka_topic`

## [4.24.0] - 2024-08-21

- Recursively mark attributes of sensitive blocks as sensitive due to an [issue in Terraform](https://github.com/hashicorp/terraform-plugin-sdk/issues/201)

## [4.23.0] - 2024-08-19

- Use enum choices from [code-generated client](https://github.com/aiven/go-client-codegen)
- Add missing enum choices to resource descriptions: `aiven_project_user` and `aiven_mysql_user`
- Mark more `aiven_grafana`, `aiven_opensearch` and `aiven_service_integration_endpoint` sensitive fields

## [4.22.0] - 2024-08-02

- Show "create" plan for missing `aiven_kafka_schema`
- Update to the latest user configuration settings [b4f7493](https://github.com/aiven/terraform-provider-aiven/commit/b4f7493fd642dfd5eb96ba3768a5d1fbebd8af78)

## [4.21.0] - 2024-07-23

- Fix `aiven_transit_gateway_vpc_attachment`: remove `peer_region` deprecation, mark the field as create only.
- Add `aiven_valkey` resource as a beta resource
- Add `aiven_valkey_user` resource as a beta resource
- Temporarily remove the `aiven_project_user` deprecation until we find a suitable alternative.

## [4.20.0] - 2024-06-26

- Mark several sensitive user config fields as "sensitive"
- Fix aiven-go-client dependency version
- Fix `aiven_organization_user_group` resource - `name` field is required
- Use `TypeSet` for `config_properties_exclude`

## [4.19.1] - 2024-05-05

- Add `config_properties_exclude` attribute to `aiven_mirrormaker_replication_flow` resource

## [4.19.0] - 2024-05-28

- Add `is_super_admin` flag to `aiven_organization_application_user` resource
- Add `aiven_mirrormaker_replication_flow` replication factor
- Remove `aiven_mirrormaker_replication_flow` global mutex, the backend has been fixed
- Remove service version validation to allow running new service versions without the provider upgrade
- Fix `aiven_organization_application_user_token` crashes with empty optional fields
- Fix `ip_filter` conversion issue

## [4.18.0] - 2024-05-22

- Fix `ip_filter`, `ip_filter_string`, and `ip_filter_object` crash when receive an unexpected type
- Add missing connection info fields to various service types

## [4.17.0] - 2024-05-07

- Fix `aiven_kafka_connector` fails to create resource with 201 error
- Remove the beta flag from `aiven_organization_application_user` and `aiven_organization_application_user_token` resources
- Fix incorrect token string used in Plugin Framework version of the provider

## [4.16.0] - 2024-04-30

- Fix incorrect timeout values used in certain cases
- Fix sending `aiven_kafka_topic` config default values
- Fix sending `false` values in `aiven_kafka_topic` config
- Fix `aiven_pg` user config fields with `__dot__` substring in name
- Validate `aiven_kafka_topic` topic name conflict on `terraform plan`
- Mark service connection info blocks as `sensitive`. See SDK [bug](https://github.com/hashicorp/terraform-plugin-sdk/issues/201).
- Remove redundant service connection info fields
- Add Thanos resource (`aiven_thanos`) to allow for the creation and management of Thanos services (currently available as beta)

## [4.15.0] - 2024-03-21

- Add Dragonfly resource (`aiven_dragonfly`) to allow for the creation and management of Dragonfly services
- Use new user config generator to generate service integration and service integration endpoint configs
- Fix `aiven_kafka_schema` version update
- Add `external_aws_cloudwatch_logs`, `external_elasticsearch_logs_user_config`, `external_opensearch_logs_user_config`,
  `prometheus_user_config` service integration configs
- Fix `aiven_kafka_schema` Protobuf normalization
- Fix `aiven_service_integration_endpoint` for `external_postgresql` type
- Add `AIVEN_ALLOW_IP_FILTER_PURGE` environment variable to allow purging of IP filters. This is a safety feature to
  prevent accidental purging of IP filters, which can lead to loss of access to services. To enable purging, set the
  environment variable to any value before running Terraform commands.
- Use `TypeSet` for `ip_filter_object`
- Fix incorrect behavior of `aiven_mirrormaker_replication_flow` schema fields:
  - `sync_group_offsets_enabled`
  - `sync_group_offsets_interval_seconds`
  - `emit_backward_heartbeats_enabled`
  - `offset_syncs_topic_location`
  - `replication_policy_class`
- Remove the beta flag for `aiven_organization_user_group_member` and `aiven_organization_group_project` resources

## [4.14.0] - 2024-02-20

- Refactor deprecated `resource.StateChangeConf`/`resource.StateRefreshFunc` usage to their equivalent with `retry`
- Fix `aiven_kafka_schema` JSON object diff suppress function
- Use new user config generator to generate service configs
- Use `TypeSet` for `ip_filter`, `ip_filter_string` fields
- Fix `aiven_organization_user_group` resource - `description` field is required
- Use golang 1.22
- Output explicitly `termination_protection = true -> false` when service property is removed
- Fix `aiven_flink_application_deployment` deletion

## [4.13.3] - 2024-01-29

- Update example code in flink_application_version docs to be valid and not use deprecated schema members
- Fix `aiven_organization_user_group_member` deletion

## [4.13.2] - 2024-01-25

- Fix `aiven_organization_group_member` fill model
- Add `aiven_flink` service deletiong check

## [4.13.1] - 2024-01-19

- Fix `aiven_organization_user` datasource get method

## [4.13.0] - 2024-01-12

- Add organization application users support
- Add organization application user tokens support
- Configure "insufficient broker" error retries timeout
- Enable `local_retention_*` fields in `aiven_kafka_topic` resource
- Validate that `local_retention_bytes` is not bigger than `retention_bytes`

## [4.12.1] - 2024-01-05

- Temporarily disable `local_retention_*` fields in `aiven_kafka_topic` resource

## [4.12.0] - 2024-01-03

- Fix insufficient brokers error when create kafka topic
- Tiered storage support for `aiven_kafka_topic` resource

## [4.11.0] - 2024-01-02

- Add support for `aiven_organization_group_project`. Please note that this resource is in the beta stage, and to use it, you would need to set the environment variable PROVIDER_AIVEN_ENABLE_BETA to a non-zero value.
- Deprecating `aiven_organization_user` resource and update data source logic that will be used instead of the corresponding resource

## [4.10.0] - 2023-12-27

- Deprecating `project_user`, `account_team` and `account_team_member` resources
- Deprecate `account_authentication` resource. Resource creation is removed, while other operations are still possible.
- Fix unmarshalling empty userconfig crash
- Never skip basic auth username/password in service integrations user config options when sending them to the API
- Add `emit_backward_heartbeats_enabled` field support in Mirrormaker replication flow
- Add validation for email fields in `account_team_member`, `organization_user` and `project_user` resources to check
  if email is lowercase and valid
- Add `connection_uri` field to components of all services. This field contains connection information for the
  component, and is a combination of the `host` and `port` fields
- Add `external_postgresql` and `external_google_cloud_bigquery` service integration endpoints
- Do not return error on `aiven_account_team_member` deletion if the member does not exist
- Deprecating `aiven_organization_user` resource and update data source logic that will be used instead of the corresponding resource
- Retry kafka topic creation error `Cluster only has N broker(s), cannot set replication factor to M`
- Fix Kafka Topic migration issues from V3 to V4.
- Fix V3 to V4 migration issue related to cloud_name diff.
- Add support for the `aiven_organization_user_group_member` resource, allowing the association of groups with the users. Please note that this resource is in the beta stage, and to use it, you would need to set the environment variable PROVIDER_AIVEN_ENABLE_BETA to a non-zero value.

## [4.9.4] - 2023-12-13

- Fix race issues with `aiven_mirrormaker_replication_flow` on create/update/delete operations
- Add `tech_emails` to services

## [4.9.3] - 2023-10-27

- Deprecating `project_user`, `account_team` and `account_team_member` resources
- Fix incorrect read context in MySQL user resource
- Improve Kafka Topic creation and read performance

## [4.9.2] - 2023-10-04

- Fix `aiven_organization_user` created time serialization

## [4.9.1] - 2023-10-03

- Allow to modify `pg_user` replication settings
- Fix `aiven_project_user` 409 error handling
- Allow usage of `protobuf` schema in Kafka

## [4.9.0] - 2023-09-18

- Fix IP Filter migrations error
- Fix `aiven_transit_gateway_vpc_attachment` resource update
- Fix service IP filters normalization
- Fix improper omitting in `ToAPI`
- Add OpenSearch Security Plugin support (`aiven_opensearch_security_plugin_config` resource)
- Deprecate `add_account_owners_admin_access` and `use_source_project_billing_group` fields in `aiven_project` resource

## [4.8.1] - 2023-08-23

- Add Organization User Groups support
- Fixed incorrect `account_id` behavior in mixed constraint setup in `aiven_project` resource
- Use updated aiven-go-client with enhanced retries
- Change `plan` from optional to required
- Improve `disk_space` deprecation message to become more explicit to migrating users
- Fix account deletion flakiness
- Make `aiven_organization` resource and data source use Terraform Plugin Framework
- Add `scheduled_rebalance_max_delay_ms` property to `aiven_kafka` and `aiven_kafka_connect`
- Add `openid` property to `aiven_opensearch`
- Add `auto_login` property to `aiven_grafana`
- Add GCP private service connect (Privatelink) support

## [4.7.0] - 2023-07-11

- Add `parent_id` field in `aiven_project` and `aiven_billing_group` resources
  - This field accepts either organization or account ID
- Deprecated `account_id` in `aiven_project` and `aiven_billing_group` resources
  - Please use `parent_id` instead, `account_id` is going to be removed in the next major release
- Fix `parent_id` storing mechanism in `aiven_organizational_unit`
- Fix `aiven_kafka_schema` import produced non-empty plan for `compatibility_level` attribute
- Add `aiven_organization_user` resource and data source

## [4.6.0] - 2023-06-28

- Fix wrong ID used for `organization` resources
- Add `aiven_flink_application_deployment` resource

## [4.5.0] - 2023-06-14

- Fix not being able to be set `ip_filter` and similar array fields in user config options to an empty array
- `aiven_kafka_topic` field `unclean_leader_election_enable` is deprecated
- Fix CIDRs handled improperly in VPC resources
- Deprecate `peer_region` field of `aiven_transit_gateway_vpc_attachment` resource
- Un-deprecated `aiven_account_team_project`, it will be deprecated when there is an alternative

## [4.4.1] - 2023-06-01

- Suppress diff for `schema_type` on `aiven_kafka_schema` resource import
- Improve Kafka Topic 404 error handling
- Use default validation.StringInSlice

## [4.4.0] - 2023-05-23

- Set `TypeSet` for `user_peer_network_cidrs` field
- Added `aiven_organization` and `aiven_organizational_unit` resources and corresponding data sources
- Deprecated the `aiven_account` resource, added a hint for the following fields that won't be supported in the future:
  - `account_id` (should use `id` instead)
  - `primary_billing_group_id`
  - `owner_team_id`
  - `is_account_owner`
- Deprecated `aiven_account_team_project` resource

## [4.3.0] - 2023-05-10

- Added docs and validation for `aiven_service_integration_endpoint`
- Dropped `signalfx` from supported integration types
- Fix MySQL user creation authentication field
- Fix Account SAML Field mapping set method
- Adjust generated SQL for ClickHouse privilege grants
- Fix `required` not generated for top level fields for user config options
- Deprecate `karapace` option on `aiven_kafka` schema
- Deprecate service `disk_space` field
- Fix `required` properties not always sent to the API for user config options

## [4.2.1] - 2023-04-06

- Add GCP virtual network peering example
- Fix add conflicting with logic for M3DB `namespaces` and `ip_filters`
- Fix double apply necessity when migrating from `ip_filter` to `ip_filter_object` and similar fields

## [4.2.0] - 2023-03-29

- Add `aiven_m3db` specific configuration options
- Fix `aiven_kafka_topic`: add client-side validation for the `partitions` field
- Make `config` field of `aiven_kafka_connector` resource non-sensitive
- Add string-suffixed alias fields for `ip_filter` and `namespaces` user config options

## [4.1.3] - 2023-03-22

- Fix the provider not working completely due to incorrect Plugin Framework configuration

## [4.1.2] - 2023-03-21

- Fix for "to API" converter for objects and arrays

## [4.1.1] - 2023-03-15

- Fix `aiven_kafka_topic` create. Now conflicts if topic exists

## [4.1.0] - 2023-03-07

- Fix `add_account_owners_admin_access` type issue
- Add service integration `integration_type` enum validation
- Rename `aiven_flink_application_version` field `sources` and `sinks` to `source` and `sink`

## [4.0.0] - 2023-02-24

- Mark `ip_filter` and `namespaces` user configuration options as deprecated
- Make schema fields use strict types instead of string
- Add support for strict types in diff functions
- Add `stateupgrader` package which aims to assist with upgrading from v3.x to v4.0.0
- Remove deprecated resources (with data sources): `aiven_database`, `aiven_service_user`, `aiven_vpc_peering_connection`
- Remove deprecated resources: `aiven_flink_table` and `aiven_flink_job`
- Switch to Terraform plugin framework
- Add support for Terraform protocol version 6

## [3.12.1] - 2023-02-16

- Fix `CreateOnlyDiffSuppressFunc`
- Fix `maintenance_window_dow` set `never` blocks resource update
- Fix Kafka `default_acl` field acting incorrectly on import and creation

## [3.12.0] - 2023-02-10

- Fix user config serialization with null values only
- Fix incorrect state name spelling in Flink resource
- Add `aiven_flink_application` resource
- Add `aiven_flink_application_version` resource
- Add `external_aws_cloudwatch_metrics` integration type

## [3.11.0] - 2023-01-09

- Fix races in tests
- Add support for normalization of `ip_filter_object` user config options
- Improve team member deletion
- Recreate missing kafka topics

## [3.10.0] - 2022-12-14

- Add ClickHouse examples:
  - Standalone service
  - Integration with Kafka source
  - Integration with PostgreSQL source
- Fix VPC peering ID parser
- Add `offset_syncs_topic_location` support for `aiven_mirrormaker_replication_flow` resource
- Add `ssl` and `kafka_authentication_method` output support in service components
- Fix `admin_username` and `admin_password` fields diff

## [3.9.0] - 2022-12-01

- Add static schema generator for user config options
- Add `ip_filter_object` and `namespaces_object` user config options which are meant to extend the existing `ip_filter` and `namespaces` ones
- Revert `datasource_project_vpc` `cloud_name` and `project` deprecations
- Add extra timeout for `kafka_connect` service integration create
- Support `clickhouse_kafka` integration type in `aiven_service_integration`
- Fix `aiven_transit_gateway_vpc_attachment` fails to parse ID
- Prevent generation of `Default` field in static schema generator
- Add `self_link` field to `aiven_gcp_vpc_peering_connection` resource
- Support `clickhouse_postgresql_user_config` integration type in `aiven_service_integration`

## [3.8.1] - 2022-11-10

- Fix `GetServiceUserValidateFunc`
- Fix provider panics on `terraform import` with invalid vpc peering id
- Fix Azure vpc peering connection import
- Fix project create/update with `add_account_owners_admin_access` field
- Add OpenSearch external integration endpoint
- Add `aiven_pg_user` import example to docs
- Extend converter for the service user configuration options `ip_filter` object format
- Fix `aiven_azure_privatelink_connection_approval` import

## [3.8.0] - 2022-09-30

- Fix `aiven_gcp_vpc_peering_connection` creation
- Improve static IP error handling end messaging
- Fix `aiven_account_authentication` resource update, add tests
- Change `aiven_project_vpc` datasource behaviour
- Fix `aiven_service_component` optional parameters filters
- Return mirrormaker user config option
- Update user config options
- Add a converter for the service user configuration options `ip_filter` object format
- Fix the Kafka service `default_acl` criteria for the deletion of default ACLs
- Don't send empty `additional_backup_regions` to the API
- Update available ClickHouse service integrations

## [3.7.0] - 2022-09-30

- Update to the official name Aiven Provider for Terraform
- Replace older links from help.aiven.io to docs.aiven.io
- Change Flink version to 1.15
- Fix empty `user_config` and empty `ip_filters`
- Fix `tools/` consistency
- Add support for `is_account_owner` field
- Forcefully send user_config on service integration update, add `datadog_user_config`
- Add user config options schema generation
- Drop Kafka Mirrormaker 1 support as it is no longer supported by the API and was broken
- Add missing examples
- `static_ips` schema type change from list to set
- Add postgres example test
- Add cron job to run examples tests

## [3.6.0] - 2022-08-31

- Add additional disk space support
- Update disk space on refresh
- Remove the `retention_hours` logic from the `aiven_kafka_topic` resource

## [3.5.1] - 2022-08-16

- Add service disk space custom diff 404 error handling
- Fix VPC peering connection custom diff

## [3.5.0] - 2022-08-10

- Add custom diff for all types of VPC peering connections that check if a VPC connection already exists before creation
- Add error handling for service `project_vpc_id` field
- Fix version `ldflag`
- Beautify and rework `Makefile`
- Add `make` targets `build` and `build-dev`
- Add sweeper for account authentication
- Make use of `BUILD_DEV_DIR` in `Makefile`
- Use Changelog Enforcer GitHub Action
- Security fix for GitHub Actions workflows
- Fix `Makefile` acceptance tests command
- Fix Changelog Enforcer GitHub Actions workflow
- Make `Makefile` variables env changeable
- Add `dependabot.yml`
- Prevent Changelog Enforcer GitHub Actions workflow from triggering for PRs from `dependabot[bot]`
- Add `no changelog` label check in `changelog-enforcer.yml`
- Add Dependency Review workflow
- Update Changelog Enforcer workflow
- Add CodeQL workflow
- Add `opensearch_index` support to `aiven_flink_table`
- Add not found checks to the Kafka availability waiter
- Add PostgreSQL max connections and PgBouncer outputs
- Perform general code clean-up and add `revive` linter
- Add support for new user configuration options

## [3.4.0] - 2022-07-26

- Small static IP import fix
- Add acceptance test for validating 404 error handling during import
- Disable `fail-fast` on acceptance tests
- Replaced every `schema.Resource.Importer.StateContext` to `schema.ImportStatePassthroughContext`
- Got rid of all unnecessary `d.SetId("")` calls
- Replaced `vpc.parsePeeringVPCId` with `schemautil.SplitResourceID`
- Made `schemautil.SplitResourceID` throw an error when the resulting amount of parts is not equal to expected
- Marked deprecated resources deprecated
- Dropped deprecated resources from sample project
- Added *.terraform.lock.hcl to .gitignore
- Update account authentication SAML fields
- Add Flink SQL validation
- Add outputs example to the sample project

## [3.3.1] - 2022-07-15

- Fix mark user config of `aiven_kafka_connector` as sensitive as it may contain credentials
- Kafka Topic availability waiter optimization
- Fix `aiven_billing_group` datasource
- Build and use go 1.18

## [3.3.0] - 2022-07-14

- Fix auto generated documentation by bumping tfplugindocs to latest version
- Fix typos in docs and examples
- Minor acceptance tests updates
- Update the 404 error handling behavior during import
- Use SDKv2 `schema.ImportStatePassthroughContext` as the importer state function
- Add Kafka `aiven_kafka_user.username` validation similar to Kafka ACL resource
- Add scheduled CI sweep job
- Add acceptance test for modifying service's user config
- Add support for `auto_join_team_id` in account authentication resource
- Fix PostgreSQL acceptance test with `static_ips` to actually check for their existence after service's creation
- Add acceptance test coverage for modification of `static_ips` in Terraform configs (via PostgreSQL)
- Fix `CustomizeDiffCheckStaticIpDisassociation` behavior
- Made it possible to delete static IPs in a single step, without dissociating them
- Fix typo in sweeper
- Fix acceptance test `TestAccAivenKafkaACL_basic`
- Add support for Kafka Schema Registry Access Control Lists resource
- Fix release actions
- Build with go 1.17

## [3.2.1] - 2022-06-29

- Fix documentation for M3DB namespaces and other documentation and examples improvements
- Fix `aiven_service_integration` poke the Kafka connect API to ensure the creation of subsequent connectors
- Change acceptance tests Terraform formating with `katbyte/terrafmt`
- Add issue and pull request templates
- Add community-related documentation
- Fix Kafka Connector's `config.name` validation
- Change acceptance tests Kafka service plan from `business-4` to `startup-2`
- Fix VPC peering connection import
- Add the CI sweep feature and rework the GitHub CI pipeline
- Refine datasource service component error message
- Fix Redis service creation when persistence is off
- Allow retrieving Project VPC data-source by ID

## [3.2.0] - 2022-06-21

- Fix typos in documentation and examples
- Fix Redis service creation when persistence is off
- Allow retrieving project VPC by ID

## [3.1.0] - 2022-06-13

- Add Kafka schema JSON support
- Add support for new `aiven_flink_table` fields
- Expose `aiven_kafka_acl` internal Aiven ID
- Fix `aiven_project` creation handling, if a project exists, then error if trying to create it again
- Add copy from billing group support
- Add service tags support
- Add project tags support
- Fix typos and errors in documentation and examples

## [3.0.0] - 2022-05-13

- `aiven_service` and `aiven_elasticsearch` resources were deleted
- `aiven_project` resource previously deprecated schema field were deleted
Deprecated resources and data-sources:
- `aiven_database`
- `aiven_service_user`
- `aiven_vpc_peering_connection`
New resources and data-sources:
- `aiven_aws_vpc_peering_connection`
- `aiven_azure_vpc_peering_connection`
- `aiven_gcp_vpc_peering_connection`
- `aiven_influxdb_user`
- `aiven_influxdb_database`
- `aiven_mysql_user`
- `aiven_mysql_database`
- `aiven_redis_user`
- `aiven_pg_user`
- `aiven_pg_database`
- `aiven_cassandra_user`
- `aiven_m3db_user`
- `aiven_m3db_user`
- `aiven_opensearch_user`
- `aiven_kafka_user`
- `aiven_clickhouse_user`
- `aiven_clickhouse_database`

## [2.7.3] - 2022-05-02

- Add missing user configuration option `thread_pool_index_size`

## [2.7.2] - 2022-04-22

- Add support for new user configuration options
- Add support for `primary_billing_group_id` to account
- Fix project resource import and read for deprecated billing group fields
- Update project resource creation such that the default billing group wouldn't be created

## [2.7.1] - 2022-04-04

- Fix account team member deletion
- Remove Elasticsearch acceptance tests
- Fix missing kafka service username and password fields

## [2.7.0] - 2022-02-18

- Add support for `aiven_clickhouse_grant` resource
- Fix `aiven_kafka` karapace migration
- Update `aiven_kafka_connector` examples
- Fix `aiven_kafka` resource 404 handling
- Add support for `aiven_clickhouse_role`
- Minor documentation and acceptance tests updates
- Add documentation for `azure_privatelink_connection_approval` resource
- Add implement importer for `aiven_static_ip`
- Fix `aiven_flink_table` possible startup modes for kafka

## [2.6.0] - 2022-02-04

- Add provider version to user agent
- Add support for `aiven_static_ip` resource
- Add support for `aiven_azure_privatelink_connection_approval` resource
- Add support for `aiven_clickhouse`, `aiven_clickhouse_user` and `aiven_clickhouse_database` resources
- Add comment trigger for acceptance tests
- Minor changes in the layout and tooling

## [2.5.0] - 2022-01-14

- Add a new field to `aiven_service_user` resource - Postgres Allow Replication

## [2.4.3] - 2022-01-13

- Add forgotten 'disk_space_used' attribute to the deprecated service resource

## [2.4.2] - 2022-01-12

- mark service_user.password as computed again

## [2.4.1] - 2022-01-11

- Reformat embedded terraform manifests
- Disable service `disk_space` default values propagation
- Add ClickHouse service beta support
- Validation of `kafka_schema` during `terraform plan` (only for schema update, not for initial creation)
- Fix saml auth provider URL's
- `aiven_kafka_topic` resource optimizations
- Fix a typo in the account acceptance test
- Fix project creation with `account_id` empty and add possibility to dissociate project from an account by not
  setting `account_id`
- Fix typos in documentation and examples
- Add `resource_elasticsearch_acl` acceptance tests
- Improve logging for service waiter

## [2.4.0] - 2021-12-01

- Add data source support for `aiven_billing_group`
- Sync flink api
- Add support for dynamic disk sizes in service creation and updates
- Add support for Kafka service Confluent SR/REST to Karapace migration

## [2.3.2] - 2021-11-10

- Fix bug in `resource_service_integration` that would lead to configs that are doubly applied, resulting in API errors

## [2.3.1] - 2021-11-05

- Fix `aiven_transit_gateway_vpc_attachment` update operation
- Fix `ip_filter` sorting order issue
- Fix issue with importing an OS that was migrated from ES in the webconsole

## [2.3.0] - 2021-10-22

- Add Flink support that includes: `aiven_flink`, `aiven_flink_table` and `aiven_flink_job` resources
- Autogenerated documentation
- Change service's `user_config` array behaviour
- Add support for `oneOf` user configuration option type
- Add Debug mode and documentation
- Add a new field `add_account_owners_admin_access` to the `aiven_project` resource
- Add Azure Privatelink support
- Fix typo in OpenSearch resource docs

## [2.2.1] - 2021-09-24

- Add support for new `aiven_mirrormaker_replication_flow` fields
- Add `aiven_connection_pool` username field optional
- Fix invalid argument name in opensearch example

## [2.2.0] - 2021-09-21

- Split `aiven_elasticsearch_acl` into `aiven_elasticsearch_acl_config` and `aiven_elasticsearch_acl_rule`
- Deprecated `aiven_elasticsearch_acl` and `aiven_elasticsearch`
- Add Opensearch support
- Add support for new user configuration options
- Add service integration creation waiter
- Add short (card's last 4 digit) card id support to a `aiven_billing_group` resource
- Rework Aiven API 409 error handling
- Fix Opensearch and Elasticsearch index_patterns deletion
- Fix `aiven_project` billing email apply loop

## [2.1.19] - 2021-08-26

- Add code of conduct
- Improve acceptance tests and documentation
- Add none existing resource handling

## [2.1.18] - 2021-08-10

- Change service integration behaviour
- Fix vpc peering connection deletion error handling
- Add GitHub golangci-lint workflow and make codebase compliant with the new code checks
- Fix `aiven_transit_gateway_vpc_attachment` crashing issue

## [2.1.17] - 2021-07-09

- Add a new field to `aiven_service_user` resource - Redis ACL Channels

## [2.1.16] - 2021-07-01

- Add `delete_retention_ms` to `aiven_kafka_topic` resource read method
- Add `use_source_project_billing_group` support for `aiven_project` resource
- Add service integration `endpoint_id` validation
- Add VPC peering connection `state_info` field type conversion

## [2.1.15] - 2021-06-14

- Add database deletion waiter
- Remove default values for user configuration options
- Improve documentation and examples
    - Add Prometheus integration example
    - Add example for Datadog metrics integration

## [2.1.14] - 2021-05-18

- Fix kafka topic acceptance test destroy check
- Fix `aiven_project_user` read method
- Use golang 1.16
- Remove GitHub pages and supporting code
- Rework documentation and examples
    - New README file structure
    - Removed the Getting Started guide and merged its contents on `docs/index.md`
    - Splitting `docs/index.md` contents in other pages on the guides
    - In examples use data source for the Aiven Project instead of resource
    - In examples use `aiven_<svc>` resource instead of `aiven_service`

## [2.1.13] - 2021-05-07

- Resend `aiven_account_team_member` and `project_user` invitations if expired

## [2.1.12] - 2021-04-20

- Improve documentation
    - Add missing import instructions
    - Add `aiven_billing_group` documentation
    - Fix required and optional `aiven_connection_pool` options
    - Updates to `MirrorMaker` arguments list
- Fix error message for prometheus user creation
- Fix project `technical_emails` and `billing_emails` fields schema
- Add support for new user configuration options
- Add MySQL example

## [2.1.11] - 2021-04-09

- Add support for Kafka Topic Tags
- Fix project name updates
- Improve documentation and examples

## [2.1.10] - 2021-04-01

- Improve documentation and error handling
- Add support to copy billing group from source project during creation
- Kafka Topic creation and availability improvements
- Change Kafka Topic resource read logic for deprecated fields
- Add support for new user configuration options
- Add examples using external kafka schema file
- Fix account team project update function typo

## [2.1.9] - 2021-03-11

- Add support for AWS Privatelink

## [2.1.8] - 2021-03-04

- Add support for new user configuration options
- Azure fields settings for VPC peering connection refactoring
- Add example initial configuration for GCP Marketplace
- Improve documentation and error messages
- Add empty string validation for OptionalStringToX conversion

## [2.1.7] - 2021-02-11

- Fix `lc_ctype` PostgreSQL database parameter
- Minor documentation improvements

## [2.1.6] - 2021-02-02

- Add Kafka Topic graceful deletion logic
- New Kafka Topic waiter and caching logic, uses v2 endpoint when available.
- Add security policy
- Improve project refresh handling when card id is incorrect
- Uses latest Terraform SDK v2.4.2
- Minor documentation improvements

## [2.1.5] - 2021-01-21

- Add support for renaming projects (only allowed for projects with no powered on services)
- Use latest go-client which fixes Kafka Topic consumer group parsing issue
- Add support for new user configuration options
- Update documentation with the newly available user configuration options
- Kafka topic availability improvements

## [2.1.4] - 2021-01-15

- Improve Kafka Topic caching
- Billing group: change project fields behaviour

## [2.1.3] - 2021-01-11

- Add support for PG upgrade check task
- Use latest go-client version
- Add support for billing groups resource
- Add support for new service user configuration options
- Add support for new service integrations user configuration options
- Use Terraform SDK v2 function instead of deprecated
- Improve formatting of the code and remove unused functions
- Fix project vpc error handling
- Update service user and service integration documentation
- Fix account team member error handling
- Remove travis ci config

## [2.1.2] - 2020-12-09

- Change VPC peering connection state handling
- Add terraform modules related docs
- Add context support for vpc peering connection
- Add service user new password support
- Extend and improve acceptance tests
- Add project resource diff suppress logic for optional fields
- Add Kafka User Configuration options max values
- Add Redis ACL support to a `service_user` resource
- Fixed docs for mandatory kafka topic params which were marked optional
- Add service integration external user configuration settings

## [2.1.1] - 2020-11-26

- Add new `aiven_project` resource attributes
- Add MirrorMaker examples
- Add new acceptance tests and change settings
- Add support for new user configuration options and service integrations
- Add Terraform version to user agent
- Update Golang requirements
- Add support for GitHub Actions
- Improve service already exists error handling
- Fix kafka topic creation typo
- Fix float to string conversion

## [2.1.0] - 2020-11-16

- Terraform plugin sdk v2 upgrade
- Update documentation: variety of minor updates which includes fix typos / grammar
- Add `is not found` error handling for delete action for all resources
- Add `already exists` error handling for create action for all resources
- Update examples
- Optimise Kafka Topic resource performance
- Fix Kafka Topic empty config issue
- Add example initial configuration for Timescale Cloud
- Add guide for documenting issues encountered during upgrades

## [2.0.11] - 2020-10-27

- Add support for new user configuration options related to Kafka, Kafka Schema Registry, Kafka Connect, Elasticsearch
  and M3 services.

## [2.0.10] - 2020-10-23

- Fix a bug related to Kafka Topic cache layer
- Small documentation improvements

## [2.0.9] - 2020-10-22

- Add support for M3 DB and M3 Aggregator, these services are currently in beta and available
  only for selected customers; currently, components for both of these services are under development.

## [2.0.8] - 2020-10-20

- Add support for new kafka topic features
- Use go-client v1.5.10
- Improve documentation
- Add support for new user configuration options
- Simplify certain part of the code
- Fix Kafka Topic validation since value that is coming from API overflows int

## [2.0.7] - 2020-10-08

- Documentation re-formatting and enhancement
- Temporarily disable docs auto-generation
- Change Kafka Topic resource `retention_hours` behaviour according to API specification
- Use latest go-client version v1.5.9
- Use golang 1.15

## [2.0.6] - 2020-09-23

- Fix README typo in the link to the prometheus/kafka example
- Fix links for kafka schemas example
- Do not change Kafka Schema compatibility level if it is empty or omitted
- Update VPC peering connection documentation

## [2.0.5] - 2020-09-17

- Extend service integration endpoint, add user configuration options
    - `external_aws_cloudwatch_logs`
    - `external_google_cloud_logging`
    - `external_kafka`
    - `jolokia`
    - `signalfx`
- Add support for new user configuration options
- Add Azure specific behaviour for VPC peering connection resource

## [2.0.4] - 2020-09-11

- Add kafka connector read waiter
- Fix Transit Gateway VPC Attachment Azure fields issue

## [2.0.3] - 2020-09-08

- Extend VPC peering connection creation with new azure related optional fields

## [2.0.2] - 2020-09-04

- Add kafka schema subject compatibility level configuration
- Use go-client v1.5.8
- Add support for -1 user configuration options when min value can be below zero
- Update user configuration options
- Small improvements: fixed tests, formatting and documentation

## [2.0.1] - 2020-08-26

- Add support for service component `aiven_service_component` data source
- Add accounts examples and update documentation
- Add PG example and documentation
- Fix vpc `user_peer_network_cidrs` type conversion problem
- Add support for goreleaser

## [2.0.0] - 2020-08-18

- Migration to Terraform Plugin SDK 0.12
- Add transit gateway vpc attachment documentation
- Add mongo sink connector examples and tests
- Kafka ACL regex modification
- New resources:
    - `aiven_pg` PostgreSQL service
    - `aiven_cassandra` Cassandra service
    - `aiven_elasticsearch` Elasticsearch service
    - `aiven_grafana` Grafana service
    - `aiven_influxdb` Influxdb service
    - `aiven_redis` Redis service
    - `aiven_mysql` MySQL service
    - `aiven_kafka` Kafka service
    - `aiven_kafka_connect` Kafka Connect service
    - `aiven_kafka_mirrormaker` Kafka Mirrormaker 2 service

## [1.3.5] - 2020-08-11

Add support for transit gateway vpc attachment

## [1.3.4] - 2020-08-10

- Expose new azure config parameters in aiven_vpc_peering_connection resource
- Add support for new user configuration options

## [1.3.3] - 2020-08-06

Fix account team projects resource project reference bug

## [1.3.2] - 2020-07-21

- Force static compilation
- Fix user configuration options array of objects format bug
- Extend elasticsearch acceptance test
- Add support for new user configuration options

## [1.3.1] - 2020-06-30

Improve vpc_id error handling for vpc peering connection

## [1.3.0] - 2020-06-18

- Add support for Kafka Mirrormaker 2
- User go-client 1.5.5
- User service configuration options refactoring
- Fix Kafka ACL data source
- Add GitHub Pages support
- Add support for Terraform native timeouts
- Add support for Accounts Authentication
- Kafka ACL allow wildcard for username
- Replace Packr2 with go generate

## [1.2.4] - 2020-05-07

- Speed up kafka topic availability waiter
- Kafka Connect examples
- TF client timings added for the following resources:
    - aiven_vpc_peering_connection
    - aiven_project_vpc
    - aiven_service
    - aiven_kafka_topic

## [1.2.3] - 2020-03-30

Add backwards compatibility for old TF state files created before Kafka `topic` field was renamed to `topic_name`.

## [1.2.2] - 2020-03-10

- Grafana service waits until Grafana is reachable publicly (only works in case `public_access.grafana`
  configuration options is set to `true` and IP filter is set to default `0.0.0.0/0`) during resource creation or
  update.
- Project VPC resource graceful deletion.

## [1.2.1] - 2020-03-02

Terraform client-side termination protection for resources:
- aiven_kafka_topic
- aiven_database

## [1.2.0] - 2020-02-18

- Following new types of resources have been added:
    - account
    - account_team
    - account_team_member
    - account_team_project
- New configuration options
- Fix for a read-only replica service types
- Service specific acceptance tests

## [1.1.6] - 2020-02-07

Fix for a problem that appears for normal boolean user configuration settings

## [1.1.5] - 2020-02-07

Fix for a problem that appears for optional bool user configuration settings

## [1.1.4] - 2020-02-03

- Acceptance tests
- Fix <service>\_user_config population problem during import

## [1.1.3] - 2020-01-24

Service pg_user_config variant configuration option remove default value.

## [1.1.2] - 2020-01-22

- Fix for <service>\_user_config Boolean fields without default value.
- Upgrade golang version to 1.13.
- Allow the API token to be read from an env var.

## [1.1.1] - 2020-01-14

Add VPC Peering Connections `state_info` property

## [1.1.0] - 2020-01-13

Terraform support for Kafka Schema's, Kafka Connectors and Service Components

## [1.0.20] - 2020-01-03

Terraform support for Elasticsearch ACLs

## [1.0.19] - 2019-12-23

Update all service configuration attributes to match latest definitions.

## [1.0.18] - 2019-12-09

Enable Kafka topics caching and add support of the Aiven terraform plugin on Windows

## [1.0.17] - 2019-09-16

Do not explicitly wait for Kafka auxiliary services to start up in an
attempt to fix issues with Kafka service create operation timing out.

## [1.0.16] - 2019-09-02

Update all service configuration attributes to match latest
definitions.

## [1.0.15] - 2019-08-19

Switch to using packr v2

## [1.0.14] - 2019-08-16

Datasource support

## [1.0.13] - 2019-08-12

Handle API errors gracefully during Kafka topic creation.

## [1.0.12] - 2019-08-08

Always wait for VPC state to become active before treating it as created.
Mark more URIs as sensitive.

## [1.0.11] - 2019-08-06

Suppress invalid update when maintenance window is unset.
Handle lc_collate and lc_ctype database attributes better.
Report Terraform provider version to server.
Treat service_uri as sensitive to avoid showing password in plain text.
Fix importing existing aiven_database resource.
Support external Elasticsearch integration.
Update available advanced configuration options.

## [1.0.10] - 2019-06-10

Switch to using go.mod
Support cross-region VPC Peering.

## [1.0.9] - 2019-04-26

Build with CGO_ENABLED=0.

## [1.0.8] - 2019-04-03

Wait for VPC to reach active state before creating service.
Don't wait for Kafka aux services if network restrictions may apply.

## [1.0.7] - 2019-03-29

Support managing MySQL services. Update all service and integration
configuration attributes match latest definitions from server.

## [1.0.6] - 2019-03-11

Fix service_port type to make it properly available.
Use latest service info on create/update to make service specific
connection info available.

## [1.0.5] - 2019-02-05

Improve retry logic in some error cases.

## [1.0.4] - 2019-01-07

Fix service_username property for services.

## [1.0.3] - 2018-12-11

Ensure Kafka topic creation succeeds immediately after service creation.
Support setting service maintenance window.
Provide cloud provider's VPC peering connection ID (AWS only).

## [1.0.2] - 2018-11-21

Support for Prometheus integration and some new user config values.

## [1.0.1] - 2018-10-08

Support termination_protection property for services.

## [1.0.0] - 2018-09-27

Support all Aiven resource types. Also large changes to previously
supported resource types, such as full support for all user config
options.
**NOTE**: This version is not backwards compatible with older versions.
