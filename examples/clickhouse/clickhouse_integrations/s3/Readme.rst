ClickHouse S3 Integration with Named Collection
===============================================

This example demonstrates how to create an Aiven for ClickHouse® service and integrate it with AWS S3. The integration allows ClickHouse users to access S3 data directly.

Named Collection Access
-----------------------

The service integration creates a named collection for S3 access. **Critical**: Only users specified in the ``clickhouse_credentials_user_config`` grants block can access this named collection:

.. code-block:: hcl

   clickhouse_credentials_user_config {
     grants {
       user = aiven_clickhouse_user.app_user.username
     }
     grants {
       user = aiven_clickhouse_user.demo_user.username
     }
   }

This configuration grants ``app_user`` and ``demo_user`` access to the S3 named collections. Users not listed here cannot use the managed credentials, even if they have other S3 privileges.

Grant Management
----------------

The example demonstrates proper grant management for ClickHouse users using the ``aiven_clickhouse_grant`` resource (see `ClickHouse Grant documentation <https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/clickhouse_grant>`_):

- **Named collection access**: Controlled by the service integration grants block
- **READ and WRITE privileges**: Allow users to access S3 functions and named collections. Since ClickHouse 25.7 the ``S3`` privilege is a deprecated alias that ClickHouse stores as ``READ`` and ``WRITE``, so grant those directly to keep the Terraform plan stable
- **CREATE TEMPORARY TABLE**: Enables users to create temporary tables for data processing
