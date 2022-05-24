---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_influxdb_database Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The InfluxDB Database data source provides information about the existing Aiven InfluxDB Database.
---

# aiven_influxdb_database (Data Source)

The InfluxDB Database data source provides information about the existing Aiven InfluxDB Database.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **database_name** (String) The name of the service database. This property cannot be changed, doing so forces recreation of the resource.
- **project** (String) Identifies the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.
- **service_name** (String) Specifies the name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.

### Optional

- **id** (String) The ID of this resource.

### Read-Only

- **termination_protection** (Boolean) It is a Terraform client-side deletion protections, which prevents the database from being deleted by Terraform. It is recommended to enable this for any production databases containing critical data. The default value is `false`.

