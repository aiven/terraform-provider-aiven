---
layout: default
parent: Guides
title: Creating a Syslog Service Integration
nav_order: 5
---

# Creating a Syslog Service Integration

Let's say you have a project in your account and a Kafka service that is throwing all sorts of errors. The Aiven web console has logs built in, you cry! Sure, but you likely need the search power of Elasticsearch or the power of logging services, such as Datadog or Rsyslog servers.

In Terraform, we support adding these service integrations and it can be broken down into 2 steps:

1. Creating an `aiven_service_integration_endpoint` - This will actually create the Service Integration that can be found in the `Service Integration` panel in the Web Console.
2. Creating an `aiven_service_integration` - This will link the endpoint to a running Aiven service.

## Example

In this example, we already have a project (`my-proj`) and a service (`kafka-service1`) running. This means we can use the `datasource` to pull in those objects from Aiven.

Then, we define the `aiven_service_integration_endpoint`.

The important things here are:
* endpoint_type - The type of integration (e.g. `rsyslog`, `prometheus`, `rsyslog`)
* {type}_user_config - The user config contains the connection info for your endpoint, such as: URL, port, Certificates (as strings) and login info. {type} here is the endpoint_type you specified above
    * The documentation for these configs is [here](https://github.com/aiven/terraform-provider-aiven/blob/8c66aab13a6bf1c4c48a1cf1e105927dab1fb93d/aiven/templates/integration_endpoints_user_config_schema.json) but generated documentation for this is coming soon.

A sample script is below:

```
terraform {
  required_providers {
    aiven = ">1.2.4"
  }
}

variable "aiven_api_token" {
  type = string
}

provider "aiven" {
  api_token = var.aiven_api_token
}

data "aiven_project" "my_proj" {
  project = "my-proj"
}

data "aiven_service" "kfk1" {
  project = data.aiven_project.my_proj.project
  service_name = "kafka-service1"
}

resource "aiven_service_integration_endpoint" "rsys" {
   project = data.aiven_project.my_proj.project
   endpoint_name="Syslog TF Example"
   endpoint_type="rsyslog"
    rsyslog_user_config {
    	server = "log.me"
    	port = 514
    	tls = false # true requires Certs to be provided
    	format = "rfc5424"
    }
}

resource "aiven_service_integration" "rsys_int" {
    project = data.aiven_project.my_proj.project
    destination_endpoint_id = aiven_service_integration_endpoint.rsys.id
    destination_service_name = ""
    integration_type = "rsyslog"
    source_endpoint_id = ""
    source_service_name = data.aiven_service.kfk1.service_name
}
```
