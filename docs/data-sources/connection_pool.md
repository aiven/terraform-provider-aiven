# Connection Pool Data Source

The Connection Pool data source provides information about the existing Aiven Connection Pool.

## Example Usage

```hcl
data "aiven_connection_pool" "mytestpool" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    pool_name = "mypool"
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the connection pool
belongs to. They should be defined using reference as shown above to set up dependencies
correctly. These properties cannot be changed once the service is created. Doing so will
result in the connection pool being deleted and new one created instead.

* `pool_name` - (Required) is the name of the pool.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `database_name` - is the name of the database the pool connects to. This should be
defined using reference as shown above to set up dependencies correctly.

* `pool_size` - is the number of connections the pool may create towards the backend
server. This does not affect the number of incoming connections, which is always a much
larger number.

* `pool_mode` - is the mode the pool operates in (session, transaction, statement).

* `username` - is the name of the service user used to connect to the database. This should
be defined using reference as shown above to set up dependencies correctly.

* `connection_uri` - is a computed property that tells the URI for connecting to the pool.
This value cannot be set, only read.

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<pool_name>`