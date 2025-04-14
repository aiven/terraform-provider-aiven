# Get started with Aiven Provider for Terraform

Set up your organization on Aiven by creating your first project and user group, and granting permissions using Terraform.

This example shows you how to create a basic setup for your organization on the Aiven Platform and does the following:

* grants the organization admin [role](https://aiven.io/docs/platform/concepts/permissions) to a user
* creates a project
* creates a user group
* assigns a user to the group
* grants the user group access to the project

You can grant permissions to users and add them to groups if they are already part of your organization. Users can be added in the Aiven Console either manually by
[sending them an invite](https://aiven.io/docs/platform/howto/manage-org-users), or you can [create managed users](https://aiven.io/docs/platform/concepts/managed-users) by verifying a domain and setting up an identity provider.
To get user IDs, use the `aiven_organization_user_list` data source.

## Prerequisites

* [Install Terraform](https://www.terraform.io/downloads)
* [Sign up for Aiven](https://console.aiven.io/signup?utm_source=github&utm_medium=organic&utm_campaign=devportal&utm_content=repo)
* [Create a token](https://aiven.io/docs/platform/concepts/authentication-tokens)
* Add users to your organization by [inviting them](https://aiven.io/docs/platform/howto/manage-org-users) or by [creating managed users](https://aiven.io/docs/platform/concepts/managed-users)

## Create your first Aiven resources

1. Ensure that you have Terraform v0.13.0 or higher installed. To check the version, run:

   ```sh
   $ terraform --version
   ```

   The output is similar to the following:

   ```sh
   Terraform v1.6.2
   + provider registry.terraform.io/aiven/aiven v4.9.2
   ```

2. Clone this repository.

3. Replace the `ORGANIZATION_NAME` placeholders in the `get-started.tf` file. It's recommended to use your organization name as a prefix for the project name because project names must be globally unique.

4. In the `aiven_organization_user_group_member` resource, replace `EMAIL_ADDRESS` with the email of one of your organization users.

5. Initialize Terraform:

   ```sh
   $ terraform init
   ```

   The output is similar to the following:

   ```sh

   Initializing the backend...

   Initializing provider plugins...
   - Finding aiven/aiven versions matching ">= 4.0.0, < 5.0.0"...
   - Installing aiven/aiven v4.9.2...
   - Installed aiven/aiven v4.9.2
   ...
   Terraform has been successfully initialized!
   ...
   ```

6. To create an execution plan and preview the changes that will be made, run:

   ```sh
   $ terraform plan

   ```

7. To deploy your changes, run:

   ```sh
   $ terraform apply
   ```

   The output will be similar to the following:
   ```sh

   Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
     + create

   Terraform will perform the following actions:

   # aiven_organization_permission.example_org_permissions will be created
   + resource "aiven_organization_permission" "example_org_permissions" {
   ...
   Plan: 5 to add, 0 to change, 0 to destroy.
   ```

8. Enter yes to confirm. The output will be similar to the following:

   ```sh
   Do you want to perform these actions?
     Terraform will perform the actions described above.
     Only 'yes' will be accepted to approve.

     Enter a value: yes

   aiven_organization_permission.example_org_permissions: Creating...
   ...
   Apply complete! Resources: 5 added, 0 changed, 0 destroyed.
   ```

## Verify the changes in the Aiven Console

Log into the [Aiven Console](https://console.aiven.io/) to see the changes you made in your organization.

### View your project and user group

1. In the organization, click **Projects** and select your project.

2. Click **Permissions** to see the user group you granted permissions to.

### View group details

1. Click **Admin**.

2. Click **Groups**.

3. Select the user group to see more information, including the members of the group.

### View organization permissions

To see the user that you granted the organization admin role to:

1. Click **Admin**.

2. Click **Permissions**.

## Clean up

To delete the example project and user group, and remove all permissions:

1. To preview the changes first, run:

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

   The output will be similar to the following:

   ```sh
   ...
   aiven_organization_user_group_member.group-members: Destroying...
   ...
   Destroy complete! Resources: 4 destroyed.
   ```

## Next steps

* Learn about organizing your resources with [organizations, units, and projects](https://aiven.io/docs/platform/concepts/orgs-units-projects).
* Create [application users](https://aiven.io/docs/platform/concepts/application-users) and tokens for more secure access to the Aiven Platform through Terraform.
