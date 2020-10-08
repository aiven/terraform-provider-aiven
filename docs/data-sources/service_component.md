# Service Component Data Source

The Service Component data source provides information about the existing Aiven service Component.

Service components can be defined to get the connection info for specific service. 
Services may support multiple different access routes (VPC peering and public access), 
have additional components or support various authentication methods. Each of these 
may be represented by different DNS name or TCP port and the specific component to 
match can be selected by specifying appropriate filters as shown below.

## Example Usage

```hcl
data "aiven_service_component" "sc1" {
    project = aiven_kafka.project1.project
    service_name = aiven_kafka.service1.service_name
    component = "kafka"
    route = "dynamic"
    kafka_authentication_method = "certificate"
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the service component
belongs to.

* `component` - (Required) is a service component name. Component may match the name of the service 
(`cassandra`, `elasticsearch`, `grafana`, `influxdb`, `kafka`, `kafka_connect`, `mysql`, 
`pg` and `redis`), in which case the connection info of the service itself is returned. 
Some service types support additional service specific components like `kibana` for 
Elasticsearch, `kafka_connect`, `kafka_rest` and `schema_registry` for Kafka, and 
`pgbouncer` for PostgreSQL. Most service types also support `prometheus`.

* `route` - (Required) is network access route. The route may be one of `dynamic`, `public`, and `private`. 
Usually, you'll want to use `dynamic`, which for services that are not in a private network 
identifies the regular public DNS name of the service and for services in a private network  
the private DNS name. If the service is in a private network but has also public access  
enabled the `public` route type can be used to get the public DNS name of the service. The  
`private` option should typically not be used.

* `kafka_authentication_method` - (Required) is Kafka authentication method. This is a value specific 
to the 'kafka' service components. And has the following available options: `certificate` 
and `sasl`. If not set by the user only entries with empty `kafka_authentication_method` 
will be selected.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `host` - is DNS name for connecting to the service component.

* `port` - is port number for connecting to the service component.

* `ssl` - whether the endpoint is encrypted or accepts plaintext. By default endpoints are
always encrypted and this property is only included for service components they may
disable encryption. If not set by the user only entries with empty `ssl` or `ssl` set 
to true will be selected.

* `usage` - is DNS usage name, and can be one of `primary`, `replica` or `syncing`. `replica` 
is used by services that have separate master and standby roles for which it identifies 
the `replica` DNS name. `syncing` is used by limited set of services to expose nodes 
before they have finished restoring state but may already be partially available, for 
example a PostgreSQL node that is streaming WAL segments from backup or current master 
but hasn't yet fully caught up.