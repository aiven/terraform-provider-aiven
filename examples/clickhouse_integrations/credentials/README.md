# ClickHouse managed credentials

This example creates a ClickHouse cluster and integrates it with credentials stored as service integration endpoints.

An S3 bucket is defined by Terraform variables and creates an service integration endpoint resource. 
The endpoint is made available to the ClickHouse cluster through a managed credentials integration.

Example PostgreSQL, MySQL and ClickHouse services are created to show how remote databases would be integrated. 
Managed credentials allow ClickHouse users to access them with the PostgreSQL & MySQL table engines, and the `remoteSecure` function.

## Prerequisites

* [Install Terraform](https://www.terraform.io/downloads)
* [Sign up for Aiven](https://console.aiven.io/signup?utm_source=github&utm_medium=organic&utm_campaign=devportal&utm_content=repo)
* [Create an authentication token](https://docs.aiven.io/docs/platform/howto/create_authentication_token.html)

## Create the example resources

1. Ensure that you have Terraform v0.13.0 or higher installed. To check the version, run:

    ```sh
    $ terraform --version 
    ```

    The output is similar to the following:

    ```sh
    Terraform version: 1.9.5
    + provider registry.terraform.io/aiven/aiven v4.9.2
    ```

2. Clone this repository.

3. Replace the placeholders in the `get-started.tf` file. It's recommended to use your organization name as a prefix for the project name.

4. Initialize Terraform:

```sh
$ terraform init
```

The output is similar to the following:

```sh

Initializing the backend...

Initializing provider plugins...
- Finding aiven/aiven versions matching ">= 4.0.0, < 5.0.0"...
- Installing aiven/aiven v4.25.0...
- Installed aiven/aiven v4.25.0
...
Terraform has been successfully initialized!
...
```

5. To create an execution plan and preview the changes that will be made, run:

```sh
$ terraform plan

```

6. To deploy your changes, run the following and enter yes to confirm:

```sh
$ terraform apply 
```

## Verify the changes in the Aiven Console

You can see the services, integration endpoints and managed credentials integrations in the [Aiven Console](https://console.aiven.io/):

1. In the project services list, you can see the ClickHouse, PostgreSQL, and MySQL services.

2. In the Integration endpoints list, you can see the S3 endpoint and the external ClickHouse, PostgreSQL, and MySQL endpoints.

3. In the overview page for the `clickhouse-gcp-eu` service, the managed credentials are show in the Data Integrations section and the Data pipeline view shows the four integrations as active.

## Clean up

To delete the example resources: 

1. To preview the changes first, run:

```sh
$ terraform plan -destroy 
```

The output shows what changes will be made when you run the `destroy` command.

2. To delete all resources, run:

```sh
$ terraform destroy 
```

3. Enter yes to confirm the changes:

```sh
Plan: 0 to add, 0 to change, 13 to destroy
...

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes
```
