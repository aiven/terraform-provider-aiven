# Automatically scale disk storage for your Aiven services

[Disk autoscaler](https://aiven.io/docs/platform/howto/disk-autoscaler#disable-disk-autoscaler) automatically increases
the storage capacity of your Aiven service when it's running out of space. Disk autoscaler only increases storage, it doesn't scale down.

To enable disk autoscaling, you create an autoscaler integration endpoint and add the integration to your service. This example:

* creates an Aiven for PostgreSQL® service
* creates an autoscaler integration endpoint with a maxiumum storage of 200GiB
* enables disk autoscaling for the service using the service integration

## Create a PostgreSQL service with disk autoscaling

1. Ensure that you have Terraform v0.13.0 or higher installed. To check the version, run:  

   ```sh
   $ terraform --version 
   ```

2. Clone this repository.

3. Add a `terraform.tfvars` file with values for your Aiven project name and your [personal token or anapplication token](https://aiven.io/docs/platform/concepts/authentication-tokens).

    ```hcl
    aiven_token         = "TOKEN"
    aiven_project_name  = "PROJECT_NAME"
    ```

4. Initialize Terraform:

   ```sh
   $ terraform init
   ```

5. To create an execution plan and preview the changes that will be made, run:

   ```sh
   $ terraform plan

   ```

6. To deploy your changes, run:

   ```sh
   $ terraform apply --auto-approve
   ```

## Verify the changes in the Aiven Console

To view your Aiven for PostgreSQL service and integration:

1. In the [Aiven Console](https://console.aiven.io), go to your project.
2. Click the name of the service to open it.
3. To view the autoscaler integration, click **Integrations**.

To view the autoscaler endpoint:

1. In the [Aiven Console](https://console.aiven.io), go to your project.
2. Click **Integration endpoints**.
3. Click **Aiven Autoscaler**.

## Clean up

To delete the service, integration, and integration endpoint:

1. Preview the changes first by running:

   ```sh
   $ terraform plan -destroy 
   ```

2. To delete all resources, run:

   ```sh
   $ terraform destroy 
   ```

3. Enter yes to confirm the changes:

   ```sh
   Plan: 0 to add, 0 to change, 4 to destroy
   ...

   Do you really want to destroy all resources?
     Terraform will destroy all your managed infrastructure, as shown above.
     There is no undo. Only 'yes' will be accepted to confirm.

     Enter a value: yes
   ```
