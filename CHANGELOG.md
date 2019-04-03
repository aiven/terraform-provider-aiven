# Changelog

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
