---
title: Home
nav_exclude: true
---

# Terraform Aiven

[Terraform](https://www.terraform.io/) provider for [Aiven.io](https://aiven.io/). This
provider allows you to conveniently manage all resources for Aiven.

**A word of cautions**: While Terraform is an extremely powerful tool that can make
managing your infrastructure a breeze, great care must be taken when comparing the
changes that are about to applied to your infrastructure. When it comes to stateful
services, you cannot just re-create a resource and have it in the original state;
deleting a service deletes all data associated with it and there's often no way to
recover the data later. Whenever the Terraform plan indicates that a service, database,
topic or other such central construct is about to be deleted, something catastrophic is
quite possibly about to happen unless you're dealing with some throwaway test
environments or are deliberately retiring the service/database/topic.

There are many properties that cannot be changed after a resource is created and changing
those values later is handled by deleting the original resource and creating a new one.
These properties include such as the project a service is associated with, the name of a
service, etc. Unless the system contains no relevant data, such changes must not be
performed.

To allow mitigating this problem, the service resource supports
`termination_protection` property. It is recommended to set this property to `true`
for all production services to avoid them being accidentally deleted. With this setting
enabled service deletion, both intentional and unintentional, will fail until an explicit
update is done to change the setting to `false`. Note that while this does prevent the
service itself from being deleted, any databases, topics or such that have been configured
with Terraform can still be deleted and they will be deleted before the service itself is
attempted to be deleted so even with this setting enabled you need to be very careful
with the changes that are to be applied.

## Requirements
- [Terraform](https://www.terraform.io/downloads.html) v0.12.X or greater
- [Go](https://golang.org/doc/install) 1.14.X or greater

## Installation

```hcl-terraform
terraform {
  required_providers {
    aiven = {
      versions = [
        "2.X"
      ]
      source = "aiven.io/provider/aiven"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_api_token
}
```
Then, initialize your Terraform workspace by running `terraform init`.

### Resource Service

```
resource "aiven_service" "myservice" {
    project = "${aiven_project.myproject.project}"
    cloud_name = "google-europe-west1"
    plan = "business-8"
    service_name = "<SERVICE_NAME>"
    service_type = "pg"
    project_vpc_id = "${aiven_project_vpc.vpc_gcp_europe_west1.id}"
    termination_protection = true
    pg_user_config {
        ip_filter = ["0.0.0.0/0"]
        pg_version = "10"
    }
    
    timeouts {
        create = "20m"
        update = "15m"
    }
}
```

`project` identifies the project the service belongs to. To set up proper dependency
between the project and the service, refer to the project as shown in the above example.
Project cannot be changed later without destroying and re-creating the service.

`cloud_name` defines where the cloud provider and region where the service is hosted
in. This can be changed freely after service is created. Changing the value will trigger
a potentially lenghty migration process for the service. Format is cloud provider name
(`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider
specific region name. These are documented on each Cloud provider's own support articles,
like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and
[here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).

`plan` defines what kind of computing resources are allocated for the service. It can
be changed after creation, though there are some restrictions when going to a smaller
plan such as the new plan must have sufficient amount of disk space to store all current
data and switching to a plan with fewer nodes might not be supported. The basic plan
names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is
(roughly) the amount of memory on each node (also other attributes like number of CPUs
and amount of disk space varies but naming is based on memory). The exact options can be
seen from the Aiven web console's Create Service dialog.

`service_name` specifies the actual name of the service. The name cannot be changed
later without destroying and re-creating the service so name should be picked based on
intended service usage rather than current attributes.

`service_type` is the actual service that is being provided. Currently available
options are `cassadra`, `elasticsearch`, `grafana`, `influxdb`, `kafka`,
`pg` (PostgreSQL) and `redis`. This value cannot be changed after service creation.

`project_vpc_id` optionally specifies the VPC the service should run in. If the value
is not set the service is not run inside a VPC. When set, the value should be given as a
reference as shown above to set up dependencies correctly and the VPC must be in the same
cloud and region as the service itself. Project can be freely moved to and from VPC after
creation but doing so triggers migration to new servers so the operation can take
significant amount of time to complete if the service has a lot of data.

`termination_protection` prevents the service from being deleted. It is recommended to
set this to `true` for all production services to prevent unintentional service
deletions. This does not shield against deleting databases or topics but for services
with backups much of the content can at least be restored from backup in case accidental
deletion is done.

`timeouts` a custom client timeouts.

`x_user_config` defines service specific additional configuration options. These
options can be found from the [JSON schema description](aiven/templates/service_user_config_schema.json).

For services that support different versions the version information must be specified in
the user configuration. By the time of writing these services are Elasticsearch, Kafka
and PostgreSQL. These services should have configuration like

```
elasticsearch_user_config {
    elasticsearch_version = "6"
}
```

```
kafka_user_config {
    kafka_version = "2.0"
}
```

```
pg_user_config {
    pg_version = "10"
}
```

Some (very few) of the user configuration options have a dot (`.`) in their name.
That is not supported by Terraform so the provider converts any literal dots to the
text string `__dot__`. So if you want to set `foo.bar = "abc"` you need to instead
set `foo__dot__bar = "abc"`.

`service_(uri|host|port|username|password)` are computed properties that define the
URI for connecting to the service and the same info split into various parts. These
values cannot be set, only read.

`x` defines service specific additional computed values for connecting to the service
(where `x` is the type of the service). E.g. `elasticsearch.0.kibana_uri` specifies
the Kibana URI for Elasticsearch service (while `service_uri` at main level is the
connection URI for the actual Elasticsearch service itself). Note the need for using
`.0` when accessing the values due to Terraform's restrictions in defining nested
schematized values. These values cannot be set, only read.

`service_integrations` can be used to define service integrations that must exist
immediately upon service creation. By the time of writing the only such integration is
defining that MySQL service is a read-replica of another service. To define a read-
replica the following configuration needs to be added:

```
service_integrations {
    integration_type = "read_replica"
    source_service_name = "${aiven_service.mysourceservice.service_name}"
}
```

Making changes to the service integrations as well as removing the service integration
requires defining an explicit `aiven_service_integration` resource with the same
attributes (plus `project` and `destination_service_name` attributes); the backend
will handle creation of an existing read-replica integration as a no-op and will just
return the identifier of the existing integration.

Aiven ID format when importing existing resource: `<project_name>/<service_name>`.

### Separate Service Resources

Starting from the version 2 Aiven Provider supports separate Terraform resources for each 
service type available in Aiven Cloud. 

List of available resources:
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
 
Instructions on how to use them are similar to `aiven_service` with the exception that 
for example `aiven_kafka` contains only all the necessary configuration options related 
to this service type.

Each resource for certain service type has the following structure:
```
resource "aiven_<TYPE>" "my-service" {
    project = aiven_project.my-project.project
    cloud_name = "google-europe-west1"
    plan = "business-4"
    service_name = "my-service1"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    <TYPE>_user_config {
        ...
    
        <TYPE> {
            ...
        }
    }
}
```

### Resource Service Integration Endpoint

Service Integration Endpoint is an external endpoint that can be used as a source or
destination for a Service Integration. These are only defined for non-Aiven resources.

```
resource "aiven_service_integration_endpoint" "myendpoint" {
    project = "${aiven_project.myproject.project}"
    endpoint_name = "<ENDPOINT_NAME>"
    endpoint_type = "datadog"
    datadog_user_config {
        datadog_api_key = "<DATADOG_API_KEY>"
    }
}
```

`project` defines the project the endpoint is associated with.

`endpoint_name` is the name of the endpoint. This value has no effect beyond being used
to identify different integration endpoints.

`endpoint_type` is the type of the external service this endpoint is associated with.
By the time of writing the only available option is `datadog`.

`x_user_config` defines endpoint type specific configuration. `x` is the type of the
endpoint. The available configuration options are documented in
[this JSON file](aiven/templates/integration_endpoints_user_config_schema.json)

Aiven ID format when importing existing resource: `<project_name>/<endpoint_id>`. The
endpoint identifier (UUID) is not directly visible in the Aiven web console.

### Resource Service Integration

Service Integration defines an integration between two Aiven services or between Aiven
service and an external integration endpoint. Integration could be for example sending
metrics from Kafka service to an InfluxDB service, getting metrics from an InfluxDB
service to a Grafana service to show dashboards, sending logs from any service to
Elasticsearch, etc.

```
resource "aiven_service_integration" "myintegration" {
    project = "${aiven_project.myproject.project}"
    destination_endpoint_id = "${aiven_service_integration_endpoint.myendpoint.id}"
    destination_service_name = ""
    integration_type = "datadog"
    source_endpoint_id = ""
    source_service_name = "${aiven_service.testkafka.service_name}"
}
```

`project` defines the project the integration belongs to.

`destination_endpoint_id` or `destination_service_name` identifies the target side of
the integration. Only either endpoint identifier or service name must be specified. In
either case the target needs to be defined using the reference syntax described above to
set up the dependency correctly.

`integration_type` identifies the type of integration that is set up. Possible values
include `dashboard`, `datadog`, `logs`, `metrics` and `mirrormaker`.

`source_endpoint_id` or `source_service_name` identifies the source side of the
integration. Only either endpoint identifier or service name must be specified. In either
case the source needs to be defined using the reference syntax described above to set up
the dependency correctly.

`x_user_config` defines integration specific configuration. `x` is the type of the
integration. The available configuration options are documented in
[this JSON file](aiven/templates/integrations_user_config_schema.json). Not all integration
types have any configurable settings.

Aiven ID format when importing existing resource: `<project_name>/<integration_id>`.
The integration identifier (UUID) is not directly visible in the Aiven web console.

## Datasource options

This section describes all the different datasources supported by the provider. The options
listed describe the options required to identify the datasource, see the equivalent resource
for options that can be referenced via the datasource.

(All examples are written using datasource references.)

### Datasource Project

```
data "aiven_project" "myproject" {
    project = "<PROJECT_NAME>"
}
```

`project` defines the name of the project.

### Datasource Service

```
data "aiven_service" "myservice" {
    project = data.aiven_project.myproject.project
    service_name = "<SERVICE_NAME>"
}
```

`project` identifies the project the service belongs to.

`service_name` specifies the actual name of the service.

### Datasource Database

```
data "aiven_database" "mydatabase" {
    project = data.aiven_service.myservice.project
    service_name = data.aiven_service.myservice.service_name
    database_name = "<DATABASE_NAME>"
    termination_protection = true
}
```

`project` and `service_name` define the project and service the database belongs to.

`database_name` is the actual name of the database.

`termination_protection` is a Terraform client-side deletion protections, which prevents 
the database from being deleted by Terraform. It is recommended to enable this for any 
production databases containing critical data.

### Datasource Service User

```
data "aiven_service_user" "myserviceuser" {
    project = data.aiven_service.myservice.project
    service_name = data.aiven_service.myservice.service_name
    username = "<USERNAME>"
}
```

`project` and `service_name` define the project and service the user belongs to.

`username` is the actual name of the user account.

### Datasource Connection Pool

```
data "aiven_connection_pool" "mytestpool" {
    project = data.aiven_service.myservice.project
    service_name = data.aiven_service.myservice.service_name
    pool_name = "<POOLNAME>"
}
```

`project` and `service_name` define the project and service the connection pool
belongs to.

`pool_name` is the name of the pool.

### Datasource Kafka Topic

```
data "aiven_kafka_topic" "mytesttopic" {
    project = data.aiven_service.myservice.project
    service_name = data.aiven_service.myservice.service_name
    topic_name = "<TOPIC_NAME>"
}
```

`project` and `service_name` define the project and service the topic belongs to.

`topic_name` is the actual name of the topic account.

### Datasource Kafka ACL

```
data "aiven_kafka_acl" "mytestacl" {
    project = data.aiven_service.myservice.project
    service_name = data.aiven_service.myservice.service_name
    topic = "<TOPIC_NAME_PATTERN>"
    username = "<USERNAME_PATTERN>"
}
```

`project` and `service_name` define the project and service the ACL belongs to.

`topic` is a topic name pattern the ACL entry matches to.

`username` is a username pattern the ACL entry matches to.

### Datasource Service Integration Endpoint

```
data "aiven_service_integration_endpoint" "myendpoint" {
    project = data.aiven_project.myproject.project
    endpoint_name = "<ENDPOINT_NAME>"
}
```

`project` defines the project the endpoint is associated with.

`endpoint_name` is the name of the endpoint.

### Datasource Project User

```
data "aiven_project_user" "mytestuser" {
    project = data.aiven_project.myproject.project
    email = "john.doe@example.com"
}
```

`project` defines the project the user is a member of.

`email` identifies the email address of the user.

### Datasource Project VPC

```
data "aiven_project_vpc" "myvpc" {
    project = data.aiven_project.myproject.project
    cloud_name = "google-europe-west1"
}
```

`project` defines the project the VPC belongs to.

`cloud_name` defines the cloud provider and region where the vpc resides.

### Datasource VPC Peering Connection

```
data "aiven_vpc_peering_connection" "mypeeringconnection" {
    vpc_id = data.aiven_project_vpc.vpc_id
    peer_cloud_account = "<PEER_ACCOUNT_ID>"
    peer_vpc = "<PEER_VPC_ID/NAME>"
}
```

`vpc_id` is the Aiven VPC the peering connection is associated with.

`peer_cloud_account` defines the identifier of the cloud account the VPC has been
peered with.

`peer_vpc` defines the identifier or name of the remote VPC.

### Datasource Service Component
Service components can be defined to get the connection info for specific service. 
Services may support multiple different access routes (VPC peering and public access), 
have additional components or support various authentication methods. Each of these 
may be represented by different DNS name or TCP port and the specific component to 
match can be selected by specifying appropriate filters as shown below.

```hcl-terraform
data "aiven_service_component" "sc1" {
    project = aiven_kafka.project1.project
    service_name = aiven_kafka.service1.service_name
    component = "kafka"
    route = "dynamic"
    kafka_authentication_method = "certificate"
}
```

`project` and `service_name` define the project and service the service component
belongs to.

`component` is a service component name. Component may match the name of the service 
(`cassandra`, `elasticsearch`, `grafana`, `influxdb`, `kafka`, `kafka_connect`, `mysql`, 
`pg` and `redis`), in which case the connection info of the service itself is returned. 
Some service types support additional service specific components like `kibana` for 
Elasticsearch, `kafka_connect`, `kafka_rest` and `schema_registry` for Kafka, and 
`pgbouncer` for PostgreSQL. Most service types also support `prometheus`.

`route` is network access route. The route may be one of `dynamic`, `public`, and `private`. 
Usually, you'll want to use `dynamic`, which for services that are not in a private network 
identifies the regular public DNS name of the service and for services in a private network  
the private DNS name. If the service is in a private network but has also public access  
enabled the `public` route type can be used to get the public DNS name of the service. The  
`private` option should typically not be used.

`host` is DNS name for connecting to the service component.

`port` is port number for connecting to the service component.

`kafka_authentication_method` is Kafka authentication method. This is a value specific 
to the 'kafka' service components. And has the following available options: `certificate` 
and `sasl`. If not set by the user only entries with empty `kafka_authentication_method` 
will be selected.

`ssl` whether the endpoint is encrypted or accepts plaintext. By default endpoints are
always encrypted and this property is only included for service components they may
disable encryption. If not set by the user only entries with empty `ssl` or `ssl` set 
to true will be selected.

`usage` is DNS usage name, and can be one of `primary`, `replica` or `syncing`. `replica` 
is used by services that have separate master and standby roles for which it identifies 
the `replica` DNS name. `syncing` is used by limited set of services to expose nodes 
before they have finished restoring state but may already be partially available, for 
example a PostgreSQL node that is streaming WAL segments from backup or current master 
but hasn't yet fully caught up.

## Credits

The original version of the Aiven Terraform provider was written and maintained by
Jelmer Snoeck (https://github.com/jelmersnoeck).
