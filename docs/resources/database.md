# Database Resource

The Database resource allows the creation and management of Aiven Databases.

## Example Usage

```hcl
resource "aiven_database" "mydatabase" {
    project = aiven_project.myproject.project
    service_name = aiven_service.myservice.service_name
    database_name = "<DATABASE_NAME>"
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the database belongs to.
They should be defined using reference as shown above to set up dependencies correctly.

* `database_name` - (Required) is the actual name of the database.

* `lc_collate` - (Optional) default string sort order (LC_COLLATE) of the database. Default value: en_US.UTF-8.

* `lc_ctype` - (Optional) default character classification (LC_CTYPE) of the database. Default value: en_US.UTF-8.

* `termination_protection` - (Optional) It is a Terraform client-side deletion protections, which prevents the database
from being deleted by Terraform. It is recommended to enable this for any production
databases containing critical data.

None of the database properties can currently be changed after creation. Doing so will
result in the old database getting dropped and a new database created.

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<database_name>`
