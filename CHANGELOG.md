---
title: Changelog
parent: README
nav_order: 1
---# Changelog

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
 configuration options is set to `true` and IP filter is set to default `0.0.0.0/0`) during resource creation or update.
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
- Fix <service>_user_config population problem during import
 
## [1.1.3] - 2020-01-24
Service pg_user_config variant configuration option remove default value.

## [1.1.2] - 2020-01-22
- Fix for <service>_user_config Boolean fields without default value.
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
