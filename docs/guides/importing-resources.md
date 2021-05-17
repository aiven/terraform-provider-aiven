---
page_title: "Importing existing Aiven infrastructure"
---

## Importing existing Aiven infrastructure
If you have already manually created an Aiven environment, it is possible to import all resources and start managing them with Terraform.

The ID format for each resource but typically it is `<project_name>/<resource_name>` for resources that are directly under project level and `<project_name>/<service_name>/<resource_name>` for resources that belong to specific service.

As example, to import a database called `mydb` belonging to the service `myservice` in the project `myproject`, you can run:
```bash 
$ terraform import aiven_database.mydb myproject/myservice/mydb
```

In some cases the internal identifiers are not shown in the Aiven web console. In such cases the easiest way to obtain identifiers is typically to check network requests and responses with your browser's debugging tools, as the raw responses do contain the IDs.

## Using data sources
Alternatively you can define already existing, or externally created and managed, resources as [data sources](../data-sources).