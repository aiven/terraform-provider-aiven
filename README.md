# Terraform Aiven

[Terraform](https://www.terraform.io/) provider for [Aiven.io](https://aiven.io/). This
provider allows you to conveniently manage all resources for Aiven.

**A word of caution**: While Terraform is an extremely powerful tool that can make
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

## Sample project

There is a [sample project](sample.tf) which sets up a project, defines Kafka,
PostgreSQL, InfluxDB and Grafana services, one PG database and user, one Kafka topic and
user, and metrics and dashboard integration for the Kafka and PG databases.

Make sure you have a look at the [variables](terraform.tfvars.sample) and copy it over to
`terraform.tfvars` with your own settings.

## Importing existing infrastructure

All resources support importing so if you have already manually created an Aiven
environment using the web console or otherwise, it is possible to import all resources
and start managing them using Terraform. The documentation below mentions the ID format
for each resource but typically it is ``<project_name>/<resource_name>`` for resources
that are directly under project level and ``<project_name>/<service_name>/<resource_name>``
for resources that belong to specific service. E.g. to import a database called ``mydb``
belonging to service ``myservice`` in project ``myproject`` you'd do something like

```
terraform import aiven_database.mydb myproject/myservice/mydb
```

In some cases the internal identifiers are not shown in the Aiven web console. In such
cases the easiest way to obtain identifiers is typically to check network requests and
responses with your browser's debugging tools, as the raw responses do contain the IDs.

Note that as Terraform does not support creating configuration automatically you will
still need to manually create the Terraform configuration files. The import will just
match the existing resources with the ones defined in the configuration.

## Options

This section describes all the different resources that can be managed with this provider.
The list of options in this document is not comprehensive. For most part the options map
directly to [Aiven REST API](https://api.aiven.io/doc/) properties and that can be
consulted for details. For various objects called x_user_config the exact configuration
options are available in [Service User Config](templates/service_user_config_schema.json),
[Integration User Config](templates/integrations_user_config_schema.json) and in
[Integration Endpoint User Config](templates/integration_endpoints_user_config_schema.json).

### Provider

```
provider "aiven" {
    api_token = "<AIVEN_API_TOKEN>"
}
```

The Aiven provider currently only supports a single configuration option, ``api_token``.
The Aiven web console can be used to create named, never expiring API tokens that should
be used for this kind of purposes. If Terraform is used for managing existing project(s),
the API token must belong to a user with admin privileges for those project(s). For new
projects the user will be automatically granted admin role. For projects with credit card
billing this account must be the one in possession with the credit card used to pay for
the services.

### Resource Project

```
resource "aiven_project" "myproject" {
    project = "<PROJECT_NAME>"
    card_id = "<FULL_CARD_ID/LAST4_DIGITS>"
}
```

``project`` defines the name of the project. Name must be globally unique (between all
Aiven customers) and cannot be changed later without destroying and re-creating the
project, including all sub-resources.

``card_id`` is either the full card UUID or the last 4 digits of the card. As the full
UUID is not shown in the UI it is typically easier to use the last 4 digits to identify
the card. This can be omitted if ``copy_from_project`` is used to copy billing info from
another project.

``copy_from_project`` is the name of another project used to copy billing information and
some other project attributes like technical contacts from. This is mostly relevant when
an existing project has billing type set to invoice and that needs to be copied over to a
new project. (Setting billing is otherwise not allowed over the API.) This only has
effect when the project is created.

``ca_cert`` is a computed property that can be used to read the CA certificate of the
project. This is required for configuring clients that connect to certain services like
Kafka. This value cannot be set, only read.

Aiven ID format when importing existing resource: name of the project as is.

### Resource Service

```
resource "aiven_service" "myservice" {
    project = "${aiven_project.myproject.project}"
    cloud_name = "google-europe-west1"
    plan = "business-8"
    service_name = "<SERVICE_NAME>"
    service_type = "pg"
    project_vpc_id = "${aiven_project_vpc.vpc_gcp_europe_west1.id}"
    pg_user_config {
        ip_filter = ["0.0.0.0/0"]
        pg_version = "10"
    }
}
```

``project`` identifies the project the service belongs to. To set up proper dependency
between the project and the service, refer to the project as shown in the above example.
Project cannot be changed later without destroying and re-creating the service.

``cloud_name`` defines where the cloud provider and region where the service is hosted
in. This can be changed freely after service is created. Changing the value will trigger
a potentially lenghty migration process for the service. Format is cloud provider name
(``aws``, ``azure``, ``do`` ``google``, ``upcloud``, etc.), dash, and the cloud provider
specific region name. These are documented on each Cloud provider's own support articles,
like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and
[here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).

``plan`` defines what kind of computing resources are allocated for the service. It can
be changed after creation, though there are some restrictions when going to a smaller
plan such as the new plan must have sufficient amount of disk space to store all current
data and switching to a plan with fewer nodes might not be supported. The basic plan
names are ``hobbyist``, ``startup-x``, ``business-x`` and ``premium-x`` where ``x`` is
(roughly) the amount of memory on each node (also other attributes like number of CPUs
and amount of disk space varies but naming is based on memory). The exact options can be
seen from the Aiven web console's Create Service dialog.

``service_name`` specifies the actual name of the service. The name cannot be changed
later without destroying and re-creating the service so name should be picked based on
intended service usage rather than current attributes.

``service_type`` is the actual service that is being provided. Currently available
options are ``cassadra``, ``elasticsearch``, ``grafana``, ``influxdb``, ``kafka``,
 ``pg`` (PostreSQL) and ``redis``. This value cannot be changed after service creation.

``project_vpc_id`` optionally specifies the VPC the service should run in. If the value
is not set the service is not run inside a VPC. When set, the value should be given as a
reference as shown above to set up dependencies correctly and the VPC must be in the same
cloud and region as the service itself. Project can be freely moved to and from VPC after
creation but doing so triggers migration to new servers so the operation can take
significant amount of time to complete if the service has a lot of data.

``x_user_config`` defines service specific additional configuration options. These
options can be found from the [JSON schema description](templates/service_user_config_schema.json).

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


Some (very few) of the user configuration options have a dot (``.``) in their name.
That is not supported by Terraform so the provider converts any literal dots to the
text string ``__dot__``. So if you want to set ``foo.bar = "abc"`` you need to instead
set ``foo__dot__bar = "abc"``.

``service_(uri|host|port|username|password)`` are computed properties that define the
URI for connecting to the service and the same info split into various parts. These
values cannot be set, only read.

``x`` defines service specific additional computed values for connecting to the service
(where ``x`` is the type of the service). E.g. ``elasticsearch.0.kibana_uri`` specifies
the Kibana URI for Elasticsearch service (while ``service_uri`` at main level is the
connection URI for the actual Elasticsearch service itself). Note the need for using
`.0` when accessing the values due to Terraform's restrictions in defining nested
schematized values. These values cannot be set, only read.

Aiven ID format when importing existing resource: ``<project_name>/<service_name>``.

### Resource Database

```
resource "aiven_database" "mydatabase" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    database_name = "<DATABASE_NAME>"
}
```

``project`` and ``service_name`` define the project and service the database belongs to.
They should be defined using reference as shown above to set up dependencies correctly.

``database_name`` is the actual name of the database.

None of the database properties can currently be changed after creation. Doing so will
result in the old database getting dropped and a new database created.

Aiven ID format when importing existing resource: ``<project_name>/<service_name>/<database_name>``

### Resource Service User

```
resource "aiven_service_user" "myserviceuser" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    username = "<USERNAME>"
}
```

``project`` and ``service_name`` define the project and service the user belongs to.
They should be defined using reference as shown above to set up dependencies correctly.

``username`` is the actual name of the user account.

None of the service user properties can currently be changed after creation. Doing so
will result in the old database getting dropped and a new database created.

Service users have several computed properties that cannot be set, only read:

``password`` is the password of the user (not applicable for all services).

``access_cert`` is the access certificate of the user (not applicable for all services).

``access_key`` is the access key of the user (not applicable for all services).

``type`` tells whether the user is primary account or regular account.

Aiven ID format when importing existing resource: ``<project_name>/<service_name>/<username>``

### Resource Connection Pool

```
resource "aiven_connection_pool" "mytestpool" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    database_name = "${aiven_database.mydatabase.database_name}"
    pool_mode = "transaction"
    pool_name = "mypool"
    pool_size = 10
    username = "${aiven_service_user.myserviceuser.username}"
}
```

``project`` and ``service_name`` define the project and service the connection pool
belongs to. They should be defined using reference as shown above to set up dependencies
correctly. These properties cannot be changed once the service is created. Doing so will
result in the connection pool being deleted and new one created instead.

``database_name`` is the name of the database the pool connects to. This should be
defined using reference as shown above to set up dependencies correctly.

``pool_mode`` is the mode the pool operates in (session, transaction, statement).

``pool_name`` is the name of the pool.

``pool_size`` is the number of connections the pool may create towards the backend
server. This does not affect the number of incoming connections, which is always a much
larger number.

``username`` is the name of the service user used to connect to the database. This should
be defined using reference as shown above to set up dependencies correctly.

``connection_uri`` is a computed property that tells the URI for connecting to the pool.
This value cannot be set, only read.

Aiven ID format when importing existing resource: ``<project_name>/<service_name>/<pool_name>``

### Resource Kafka Topic

```
resource "aiven_kafka_topic" "mytesttopic" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    topic_name = "<TOPIC_NAME>"
    partitions = 5
    replication = 3
    retention_bytes = -1
    retention_hours = 72
    minimum_in_sync_replicas = 2
    cleanup_policy = "delete"
}
```

``project`` and ``service_name`` define the project and service the topic belongs to.
They should be defined using reference as shown above to set up dependencies correctly.
These properties cannot be changed once the service is created. Doing so will result in
the topic being deleted and new one created instead.

``topic_name`` is the actual name of the topic account. This propery cannot be changed
once the service is created. Doing so will result in the topic being deleted and new one
created instead.

Other properties should be self-explanatory. They can be changed after the topic has been
created.

Aiven ID format when importing existing resource: ``<project_name>/<service_name>/<topic_name>``

### Resource Kafka ACL

```
resource "aiven_kafka_acl" "mytestacl" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    topic = "<TOPIC_NAME_PATTERN>"
    permission = "admin"
    username = "<USERNAME_PATTERN>"
}
```

``project`` and ``service_name`` define the project and service the ACL belongs to.
They should be defined using reference as shown above to set up dependencies correctly.
These properties cannot be changed once the service is created. Doing so will result in
the topic being deleted and new one created instead.

``topic`` is a topic name pattern the ACL entry matches to.

``permission`` is the level of permission the matching users are given to the matching
topics (admin, read, readwrite, write).

``username`` is a username pattern the ACL entry matches to.

Aiven ID format when importing existing resource: ``<project_name>/<service_name>/<acl_id>``.
The ACL ID is not directly visible in the Aiven web console.

### Resource Service Integration Endpoint

Service Integration Endpoint is an external endpoint that can be used as a source or
destination for a Service Integration. These are only defined for non-Aiven resources.

```
resource "aiven_service_integration_endpoint" "myendpoint" {
    project = "${aiven_project.myproject.project}"
    endpoint_name = "<ENDPOINT_NAME>"
    endpoint_type = "datadog"
    datadog_user_config = {
        datadog_api_key = "<DATADOG_API_KEY>"
    }
}
```

``project`` defines the project the endpoint is associated with.

``endpoint_name`` is the name of the endpoint. This value has no effect beyond being used
to identify different integration endpoints.

``endpoint_type`` is the type of the external service this endpoint is associated with.
By the time of writing the only available option is ``datadog``.

``x_user_config`` defines endpoint type specific configuration. ``x`` is the type of the
endpoint. The available configuration options are documented in
[this JSON file](templates/integration_endpoints_user_config_schema.json)

Aiven ID format when importing existing resource: ``<project_name>/<endpoint_id>``. The
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

``project`` defines the project the integration belongs to.

``destination_endpoint_id`` or ``destination_service_name`` identifies the target side of
the integration. Only either endpoint identifier or service name must be specified. In
either case the target needs to be defined using the reference syntax described above to
set up the dependency correctly.

``integration_type`` identifies the type of integration that is set up. Possible values
include ``dashboard``, ``datadog``, ``logs``, ``metrics`` and ``mirrormaker``.

``source_endpoint_id`` or ``source_service_name`` identifies the source side of the
integration. Only either endpoint identifier or service name must be specified. In either
case the source needs to be defined using the reference syntax described above to set up
the dependency correctly.

``x_user_config`` defines integration specific configuration. ``x`` is the type of the
integration. The available configuration options are documented in
[this JSON file](templates/integrations_user_config_schema.json). Not all integration
types have any configurable settings.

Aiven ID format when importing existing resource: ``<project_name>/<integration_id>``.
The integration identifier (UUID) is not directly visible in the Aiven web console.

### Resource Project User

```
resource "aiven_project_user" "mytestuser" {
    project = "${aiven_project.myproject.project}"
    email = "john.doe@example.com"
    member_type = "admin"
}
```

``project`` defines the project the user is a member of.

``email`` identifies the email address of the user.

``member_type`` defines the access level the user has to the project.

Computed property ``accepted`` tells whether the user has accepted the request to join
the project; adding user to a project sends an invitation to the target user and the
actual membership is only created once the user accepts the invitation. This property
cannot be set, only read.

Aiven ID format when importing existing resource: ``<project_name>/<email>``

### Resource Project VPC

```
resource "aiven_project_vpc" "myvpc" {
    project = "${aiven_project.myproject.project}"
    cloud_name = "google-europe-west1"
    network_cidr = "192.168.0.1/24"
}
```

``project`` defines the project the VPC belongs to.

``cloud_name`` defines where the cloud provider and region where the service is hosted
in. See the Service resource for additional information.

``network_cidr`` defines the network CIDR of the VPC.

Computed property ``state`` tells the current state of the VPC. This property cannot be
set, only read.

Aiven ID format when importing existing resource: ``<project_name>/<VPC_UUID>``. The UUID
is not directly visible in the Aiven web console.

### Resource VPC Peering Connection

```
resource "aiven_vpc_peering_connection" "mypeeringconnection" {
    vpc_id = "${aiven_project.myproject.myvpc.id}"
    peer_cloud_account = "<PEER_ACCOUNT_ID>"
    peer_vpc = "<PEER_VPC_ID/NAME>"
}
```

``vpc_id`` is the Aiven VPC the peering connection is associated with.

``peer_cloud_account`` defines the identifier of the cloud account the VPC is being
peered with.

``peer_vpc`` defines the identifier or name of the remote VPC.

Computed property ``state`` tells the current state of the VPC. This property cannot be
set, only read.

Aiven ID format when importing existing resource: ``<project_name>/<VPC_UUID>/<peer_cloud_account>/<peer_vpc>``.
The UUID is not directly visible in the Aiven web console.

## Credits

The original version of the Aiven Terraform provider was written and maintained by
Jelmer Snoeck (https://github.com/jelmersnoeck).
