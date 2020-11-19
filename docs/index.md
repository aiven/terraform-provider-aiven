# Aiven Provider

This provider allows you to conveniently manage all resources for Aiven.

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

## Installation instructions for Terraform client 0.12

_If you can please upgrade your Terraform client to version 0.13 or above, in this case,
there is no need to install provider manually._

Download the Aiven provider for Linux AMD64.

```shell script
curl -Lo terraform-provider-aiven https://github.com/aiven/terraform-provider-aiven/releases/download/v2.X.X/terraform-provider-aiven-linux-amd64_v2.X.X

chmod +x terraform-provider-aiven
```

_Please specify version that you like to use and if you are not using Linux AMD64,
download the correct binary for your system from the [release page](https://github.com/aiven/terraform-provider-aiven/releases)._

Third-party provider plugins — locally installed providers, not on the registry — need to be
assigned an (arbitrary) source and placed in the appropriate subdirectory for Terraform to find and use them.
Create the appropriate subdirectory within the user plugins directory for the Aiven provider.

```shell script
mkdir -p ~/.terraform.d/plugins/aiven.io/provider/aiven/2.X/linux_amd64
```

_If you are not using Linux AMD64 replace `linux_amd64` with your `\$OS_\$ARCH`, and the same for the provider version.\_

Finally, move the Aiven provider binary into the newly created directory.

```shell script
mv terraform-provider-aiven ~/.terraform.d/plugins/aiven.io/provider/aiven/2.X/linux_amd64
```

Now Aiven provider is in your user plugins directory, you can use the provider in your Terraform configuration.

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

Then, initialize your Terraform workspace by running `terraform init`. If your Aiven provider  
is located in the correct directory, it should successfully initialize. Otherwise, move your
Aiven provider to the correct directory: `~/.terraform.d/plugins/aiven.io/provider/aiven/$VERSION/$OS_$ARCH/`.

## Example Usage

_This is only available for Terraform client 0.13 and above_

```hcl
terraform {
  required_providers {
    aiven = {
      source = "aiven/aiven"
      version = "2.X.X"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_api_token
}
```

Then, initialize your Terraform workspace by running `terraform init`.

## Sample project

There is a [sample project](https://github.com/aiven/terraform-provider-aiven/tree/master/sample.tf) which sets up a project, defines Kafka,
PostgreSQL, InfluxDB and Grafana services, one PG database and user, one Kafka topic and
user, and metrics and dashboard integration for the Kafka and PG databases.

Make sure you have a look at the [variables](https://github.com/aiven/terraform-provider-aiven/tree/master/terraform.tfvars.sample) and copy it over to
`terraform.tfvars` with your own settings.

Other examples can be found in the [examples](https://github.com/aiven/terraform-provider-aiven/tree/master/examples) folder that provides examples to:

- [Getting Started](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/getting-started.tf)
- [Account, projects, teams, and member management](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/account)
- [Elasticsearch deployment and configuration](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/elasticsearch)
- [Standalone Kafka connect deployment with custom config](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/kafka_connect)
- [Deploying Kafka with a Prometheus Service Integration](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/kafka_prometheus)
- [Deploying Kafka and Elasticsearch with a Kafka Connect Elasticsearch Sink connector](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/kafka_connectors/es_sink)
- [Deploying Kafka and Elasticsearch with a Kafka Connect Mongo Sink connector](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/kafka_connectors/mongo_sink)
- [Deploying Kafka with Schema Registry enabled and providing a schema](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/kafka_schemas)
- [Deploying Cassandra and forking (cloning the service, config and data) into a new service with a higher plan](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/cassandra_fork)
- [Deploying a Grafana service](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/service)
- [Deploying a MirrorMaker service](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/kafka_mirrormaker)
- [Deploying a MirrorMaker service with multiple global regions](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/kafka_mirrormaker_global)
- [Deploying a MirrorMaker service in an Active<=>Active confguration](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/kafka_mirrormaker_bidirectional)
- [Deploying PostgreSQL services to multiple clouds and regions](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/postgres)
- [Deploying M3 and M3 Aggregator services](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/m3)
- [Deploying on Timescale Cloud](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/timescale)

## Importing existing infrastructure

All resources support importing so if you have already manually created an Aiven
environment using the web console or otherwise, it is possible to import all resources
and start managing them using Terraform. The documentation below mentions the ID format
for each resource but typically it is `<project_name>/<resource_name>` for resources
that are directly under project level and `<project_name>/<service_name>/<resource_name>`
for resources that belong to specific service. E.g. to import a database called `mydb`
belonging to service `myservice` in project `myproject` you'd do something like

```
terraform import aiven_database.mydb myproject/myservice/mydb
```

In some cases the internal identifiers are not shown in the Aiven web console. In such
cases the easiest way to obtain identifiers is typically to check network requests and
responses with your browser's debugging tools, as the raw responses do contain the IDs.

Note that as Terraform does not support creating configuration automatically you will
still need to manually create the Terraform configuration files. The import will just
match the existing resources with the ones defined in the configuration.
	
## Using datasources

Alternatively you can define already existing, or externally created and managed, resources
as datasources. Most of the resources are also available as datasources.

## Resource options

This section describes all the different resources that can be managed with this provider.
The list of options in this document is not comprehensive. For most part the options map
directly to [Aiven REST API](https://api.aiven.io/doc/) properties and that can be
consulted for details. For various objects called x_user_config the exact configuration
options are available in [Service User Config](https://github.com/aiven/terraform-provider-aiven/tree/master/aiven/templates/service_user_config_schema.json),
[Integration User Config](https://github.com/aiven/terraform-provider-aiven/tree/master/aiven/templates/integrations_user_config_schema.json) and in
[Integration Endpoint User Config](https://github.com/aiven/terraform-provider-aiven/tree/master/aiven/templates/integration_endpoints_user_config_schema.json).

### Provider

```hcl
provider "aiven" {
    api_token = "<AIVEN_API_TOKEN>"
}
```

The Aiven provider currently only supports a single configuration option, `api_token`.
This can also be specified with the AIVEN_TOKEN shell environment variable.
The Aiven web console can be used to create named, never expiring API tokens that should
be used for this kind of purposes. If Terraform is used for managing existing project(s),
the API token must belong to a user with admin privileges for those project(s). For new
projects the user will be automatically granted admin role. For projects with credit card
billing this account must be the one in possession with the credit card used to pay for
the services.
