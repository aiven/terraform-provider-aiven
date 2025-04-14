---
page_title: "Sample Project"
---

# Sample project
There is a [sample project](https://github.com/aiven/terraform-provider-aiven/tree/master/sample_project/sample.tf) which sets up:
- An Aiven project (from a data source)
- Kafka (with a topic and user)
- PostgreSQL (with a database and user)
- InfluxDB
- Grafana (metrics and dashboard integration for the Kafka and PostgreSQL databases)

Make sure you have a look at the [variables](https://github.com/aiven/terraform-provider-aiven/tree/master/sample_project/terraform.tfvars) and add your own settings.

## Running the sample project
Run the following commands:
```bash
$ git clone https://github.com/aiven/terraform-provider-aiven.git
$ cd terraform-provider-aiven/sample_project

# add your authentication token and project name  in the terraform.tfvars file

$ terraform init
$ terraform plan
$ terraform apply
```

Now go the the web console to see all your infrastructure up and running!

## Cleaning up
Run the following commands:
```bash
$ terraform destroy
```

You can also find more examples [here](examples.md).
