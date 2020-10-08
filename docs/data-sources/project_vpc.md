# Project VPC Data Source

The Project VPC data source provides information about the existing Aiven Project VPC.

## Example Usage

```hcl
data "aiven_project_vpc" "myvpc" {
    project = "${aiven_project.myproject.project}"
    cloud_name = "google-europe-west1"
}
```

## Argument Reference

* `project` - (Required) defines the project the VPC belongs to.

* `cloud_name` - (Required) defines where the cloud provider and region where the service is hosted
in. See the Service resource for additional information.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `network_cidr` - defines the network CIDR of the VPC.

* `state` - ia a computed property that tells the current state of the VPC. This property cannot be
set, only read.

Aiven ID format when importing existing resource: `<project_name>/<VPC_UUID>`. The UUID
is not directly visible in the Aiven web console.