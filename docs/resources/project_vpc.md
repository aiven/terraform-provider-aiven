# Project VPC Resource

The Project VPC resource allows the creation and management of Aiven Project VPCs.

## Example Usage

```hcl
resource "aiven_project_vpc" "myvpc" {
    project = aiven_project.myproject.project
    cloud_name = "google-europe-west1"
    network_cidr = "192.168.0.1/24"

    timeouts {
        create = "5m"
    }
}
```

## Argument Reference

* `project` - (Required) defines the project the VPC belongs to.

* `cloud_name` - (Required) defines where the cloud provider and region where the service is hosted
in. See the Service resource for additional information.

* `network_cidr` - (Required) defines the network CIDR of the VPC.

* `timeouts` - (Required) a custom client timeouts.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `state` - ia a computed property that tells the current state of the VPC. This property cannot be
set, only read.

Aiven ID format when importing existing resource: `<project_name>/<VPC_UUID>`. The UUID
is not directly visible in the Aiven web console.