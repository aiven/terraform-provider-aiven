# Connection Pool Resource

The Connection Pool resource allows the creation and management of Aiven Connection Pools.

## Example Usage

```hcl
resource "aiven_connection_pool" "mytestpool" {
    project = aiven_project.myproject.project
    service_name = aiven_service.myservice.service_name
    database_name = aiven_database.mydatabase.database_name
    pool_mode = "transaction"
    pool_name = "mypool"
    pool_size = 10
    username = aiven_service_user.myserviceuser.username

    depends_on = [
      aiven_database.mydatabase,
    ]
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the connection pool
belongs to. They should be defined using reference as shown above to set up dependencies
correctly. These properties cannot be changed once the service is created. Doing so will
result in the connection pool being deleted and new one created instead.

* `database_name` - (Required) is the name of the database the pool connects to. This should be
defined using reference as shown above to set up dependencies correctly.

* `pool_name` - (Required) is the name of the pool.

* `username` - (Required) is the name of the service user used to connect to the database. This should
  be defined using reference as shown above to set up dependencies correctly.

* `pool_size` - (Optional) is the number of connections the pool may create towards the backend
server. This does not affect the number of incoming connections, which is always a much
larger number. The default value for this is 10.

* `pool_mode` - (Optional) is the mode the pool operates in (session, transaction, statement). The
default value for this is `transaction`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `connection_uri` - (Optional) is a computed property that tells the URI for connecting to the pool.
This value cannot be set, only read.

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<pool_name>`