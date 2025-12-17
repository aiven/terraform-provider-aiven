# Plugin Framework Migration Status

```
+-----+---------------------------------------------+--------+-------+
|   # | RESOURCE NAME                               | PLUGIN | COUNT |
+-----+---------------------------------------------+--------+-------+
|   1 | aiven_billing_group                         | yes    |     2 |
|   2 | aiven_external_identity                     | yes    |     1 |
|   3 | aiven_governance_access                     | yes    |     1 |
|   4 | aiven_organization                          | yes    |     2 |
|   5 | aiven_organization_address                  | yes    |     2 |
|   6 | aiven_organization_application_user         | yes    |     2 |
|   7 | aiven_organization_application_user_token   | yes    |     1 |
|   8 | aiven_organization_billing_group            | yes    |     2 |
|   9 | aiven_organization_billing_group_list       | yes    |     1 |
|  10 | aiven_organization_group_project            | yes    |     1 |
|  11 | aiven_organization_permission               | yes    |     1 |
|  12 | aiven_organization_project                  | yes    |     2 |
|  13 | aiven_organization_user_group_list          | yes    |     1 |
|  14 | aiven_organization_user_group_member        | yes    |     1 |
|  15 | aiven_organization_user_group_member_list   | yes    |     1 |
|  16 | aiven_organization_user_list                | yes    |     1 |
|  17 | aiven_organizational_unit                   | yes    |     2 |
|  18 | aiven_service_plan                          | yes    |     1 |
|  19 | aiven_service_plan_list                     | yes    |     1 |
|  20 | aiven_account                               |        |     2 |
|  21 | aiven_account_authentication                |        |     2 |
|  22 | aiven_account_team                          |        |     2 |
|  23 | aiven_account_team_member                   |        |     2 |
|  24 | aiven_account_team_project                  |        |     2 |
|  25 | aiven_alloydbomni                           |        |     2 |
|  26 | aiven_alloydbomni_database                  |        |     2 |
|  27 | aiven_alloydbomni_user                      |        |     2 |
|  28 | aiven_aws_org_vpc_peering_connection        |        |     2 |
|  29 | aiven_aws_privatelink                       |        |     2 |
|  30 | aiven_aws_vpc_peering_connection            |        |     2 |
|  31 | aiven_azure_org_vpc_peering_connection      |        |     2 |
|  32 | aiven_azure_privatelink                     |        |     2 |
|  33 | aiven_azure_privatelink_connection_approval |        |     1 |
|  34 | aiven_azure_vpc_peering_connection          |        |     2 |
|  35 | aiven_cassandra                             |        |     2 |
|  36 | aiven_cassandra_user                        |        |     2 |
|  37 | aiven_clickhouse                            |        |     2 |
|  38 | aiven_clickhouse_database                   |        |     2 |
|  39 | aiven_clickhouse_grant                      |        |     1 |
|  40 | aiven_clickhouse_role                       |        |     1 |
|  41 | aiven_clickhouse_user                       |        |     2 |
|  42 | aiven_connection_pool                       |        |     2 |
|  43 | aiven_dragonfly                             |        |     2 |
|  44 | aiven_flink                                 |        |     2 |
|  45 | aiven_flink_application                     |        |     2 |
|  46 | aiven_flink_application_deployment          |        |     1 |
|  47 | aiven_flink_application_version             |        |     2 |
|  48 | aiven_flink_jar_application                 |        |     1 |
|  49 | aiven_flink_jar_application_deployment      |        |     1 |
|  50 | aiven_flink_jar_application_version         |        |     1 |
|  51 | aiven_gcp_org_vpc_peering_connection        |        |     2 |
|  52 | aiven_gcp_privatelink                       |        |     2 |
|  53 | aiven_gcp_privatelink_connection_approval   |        |     1 |
|  54 | aiven_gcp_vpc_peering_connection            |        |     2 |
|  55 | aiven_grafana                               |        |     2 |
|  56 | aiven_influxdb                              |        |     2 |
|  57 | aiven_influxdb_database                     |        |     2 |
|  58 | aiven_influxdb_user                         |        |     2 |
|  59 | aiven_kafka                                 |        |     2 |
|  60 | aiven_kafka_acl                             |        |     2 |
|  61 | aiven_kafka_connect                         |        |     2 |
|  62 | aiven_kafka_connector                       |        |     2 |
|  63 | aiven_kafka_mirrormaker                     |        |     2 |
|  64 | aiven_kafka_native_acl                      |        |     1 |
|  65 | aiven_kafka_quota                           |        |     1 |
|  66 | aiven_kafka_schema                          |        |     2 |
|  67 | aiven_kafka_schema_configuration            |        |     2 |
|  68 | aiven_kafka_schema_registry_acl             |        |     2 |
|  69 | aiven_kafka_topic                           |        |     2 |
|  70 | aiven_kafka_user                            |        |     2 |
|  71 | aiven_m3aggregator                          |        |     2 |
|  72 | aiven_m3db                                  |        |     2 |
|  73 | aiven_m3db_user                             |        |     2 |
|  74 | aiven_mirrormaker_replication_flow          |        |     2 |
|  75 | aiven_mysql                                 |        |     2 |
|  76 | aiven_mysql_database                        |        |     2 |
|  77 | aiven_mysql_user                            |        |     2 |
|  78 | aiven_opensearch                            |        |     2 |
|  79 | aiven_opensearch_acl_config                 |        |     2 |
|  80 | aiven_opensearch_acl_rule                   |        |     2 |
|  81 | aiven_opensearch_security_plugin_config     |        |     2 |
|  82 | aiven_opensearch_user                       |        |     2 |
|  83 | aiven_organization_user                     |        |     2 |
|  84 | aiven_organization_user_group               |        |     2 |
|  85 | aiven_organization_vpc                      |        |     2 |
|  86 | aiven_pg                                    |        |     2 |
|  87 | aiven_pg_database                           |        |     2 |
|  88 | aiven_pg_user                               |        |     2 |
|  89 | aiven_project                               |        |     2 |
|  90 | aiven_project_user                          |        |     2 |
|  91 | aiven_project_vpc                           |        |     2 |
|  92 | aiven_redis                                 |        |     2 |
|  93 | aiven_redis_user                            |        |     2 |
|  94 | aiven_service_component                     |        |     1 |
|  95 | aiven_service_integration                   |        |     2 |
|  96 | aiven_service_integration_endpoint          |        |     2 |
|  97 | aiven_static_ip                             |        |     1 |
|  98 | aiven_thanos                                |        |     2 |
|  99 | aiven_transit_gateway_vpc_attachment        |        |     2 |
| 100 | aiven_valkey                                |        |     2 |
| 101 | aiven_valkey_user                           |        |     2 |
+-----+---------------------------------------------+--------+-------+
|     | TOTAL MIGRATED 15%                          | 26     |   178 |
+-----+---------------------------------------------+--------+-------+
```
