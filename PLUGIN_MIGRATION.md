# Plugin Framework Migration Status

```
+-----+---------------------------------------------+--------+-------+
|   # | RESOURCE NAME                               | PLUGIN | COUNT |
+-----+---------------------------------------------+--------+-------+
|   1 | aiven_account                               |        |     2 |
|   2 | aiven_account_authentication                |        |     2 |
|   3 | aiven_account_team                          |        |     2 |
|   4 | aiven_account_team_member                   |        |     2 |
|   5 | aiven_account_team_project                  |        |     2 |
|   6 | aiven_alloydbomni                           |        |     2 |
|   7 | aiven_alloydbomni_database                  |        |     2 |
|   8 | aiven_alloydbomni_user                      |        |     2 |
|   9 | aiven_aws_org_vpc_peering_connection        |        |     2 |
|  10 | aiven_aws_privatelink                       |        |     2 |
|  11 | aiven_aws_vpc_peering_connection            |        |     2 |
|  12 | aiven_azure_org_vpc_peering_connection      |        |     2 |
|  13 | aiven_azure_privatelink                     |        |     2 |
|  14 | aiven_azure_privatelink_connection_approval |        |     1 |
|  15 | aiven_azure_vpc_peering_connection          |        |     2 |
|  16 | aiven_billing_group                         | yes    |     2 |
|  17 | aiven_cassandra                             |        |     2 |
|  18 | aiven_cassandra_user                        |        |     2 |
|  19 | aiven_clickhouse                            |        |     2 |
|  20 | aiven_clickhouse_database                   | yes    |     2 |
|  21 | aiven_clickhouse_grant                      |        |     1 |
|  22 | aiven_clickhouse_role                       |        |     1 |
|  23 | aiven_clickhouse_user                       |        |     2 |
|  24 | aiven_cmk                                   | yes    |     1 |
|  25 | aiven_connection_pool                       |        |     2 |
|  26 | aiven_dragonfly                             |        |     2 |
|  27 | aiven_external_identity                     | yes    |     1 |
|  28 | aiven_flink                                 |        |     2 |
|  29 | aiven_flink_application                     |        |     2 |
|  30 | aiven_flink_application_deployment          |        |     1 |
|  31 | aiven_flink_application_version             |        |     2 |
|  32 | aiven_flink_jar_application                 |        |     1 |
|  33 | aiven_flink_jar_application_deployment      |        |     1 |
|  34 | aiven_flink_jar_application_version         |        |     1 |
|  35 | aiven_gcp_org_vpc_peering_connection        |        |     2 |
|  36 | aiven_gcp_privatelink                       |        |     2 |
|  37 | aiven_gcp_privatelink_connection_approval   |        |     1 |
|  38 | aiven_gcp_vpc_peering_connection            |        |     2 |
|  39 | aiven_governance_access                     | yes    |     1 |
|  40 | aiven_grafana                               |        |     2 |
|  41 | aiven_influxdb                              |        |     2 |
|  42 | aiven_influxdb_database                     |        |     2 |
|  43 | aiven_influxdb_user                         |        |     2 |
|  44 | aiven_kafka                                 |        |     2 |
|  45 | aiven_kafka_acl                             |        |     2 |
|  46 | aiven_kafka_connect                         |        |     2 |
|  47 | aiven_kafka_connector                       |        |     2 |
|  48 | aiven_kafka_mirrormaker                     |        |     2 |
|  49 | aiven_kafka_native_acl                      |        |     1 |
|  50 | aiven_kafka_quota                           |        |     1 |
|  51 | aiven_kafka_schema                          |        |     2 |
|  52 | aiven_kafka_schema_configuration            |        |     2 |
|  53 | aiven_kafka_schema_registry_acl             |        |     2 |
|  54 | aiven_kafka_topic                           |        |     2 |
|  55 | aiven_kafka_user                            |        |     2 |
|  56 | aiven_m3aggregator                          |        |     2 |
|  57 | aiven_m3db                                  |        |     2 |
|  58 | aiven_m3db_user                             |        |     2 |
|  59 | aiven_mirrormaker_replication_flow          |        |     2 |
|  60 | aiven_mysql                                 |        |     2 |
|  61 | aiven_mysql_database                        | yes    |     2 |
|  62 | aiven_mysql_user                            | yes    |     2 |
|  63 | aiven_opensearch                            |        |     2 |
|  64 | aiven_opensearch_acl_config                 |        |     2 |
|  65 | aiven_opensearch_acl_rule                   |        |     2 |
|  66 | aiven_opensearch_security_plugin_config     |        |     2 |
|  67 | aiven_opensearch_user                       |        |     2 |
|  68 | aiven_organization                          | yes    |     2 |
|  69 | aiven_organization_address                  | yes    |     2 |
|  70 | aiven_organization_application_user         | yes    |     2 |
|  71 | aiven_organization_application_user_token   | yes    |     1 |
|  72 | aiven_organization_billing_group            | yes    |     2 |
|  73 | aiven_organization_billing_group_list       | yes    |     1 |
|  74 | aiven_organization_group_project            | yes    |     1 |
|  75 | aiven_organization_permission               | yes    |     1 |
|  76 | aiven_organization_project                  | yes    |     2 |
|  77 | aiven_organization_user                     |        |     2 |
|  78 | aiven_organization_user_group               |        |     2 |
|  79 | aiven_organization_user_group_list          | yes    |     1 |
|  80 | aiven_organization_user_group_member        | yes    |     1 |
|  81 | aiven_organization_user_group_member_list   | yes    |     1 |
|  82 | aiven_organization_user_list                | yes    |     1 |
|  83 | aiven_organization_vpc                      |        |     2 |
|  84 | aiven_organizational_unit                   | yes    |     2 |
|  85 | aiven_pg                                    |        |     2 |
|  86 | aiven_pg_database                           | yes    |     2 |
|  87 | aiven_pg_user                               |        |     2 |
|  88 | aiven_project                               |        |     2 |
|  89 | aiven_project_user                          |        |     2 |
|  90 | aiven_project_vpc                           |        |     2 |
|  91 | aiven_redis                                 |        |     2 |
|  92 | aiven_redis_user                            |        |     2 |
|  93 | aiven_service_component                     |        |     1 |
|  94 | aiven_service_integration                   |        |     2 |
|  95 | aiven_service_integration_endpoint          |        |     2 |
|  96 | aiven_service_list                          | yes    |     1 |
|  97 | aiven_service_plan                          | yes    |     1 |
|  98 | aiven_service_plan_list                     | yes    |     1 |
|  99 | aiven_static_ip                             |        |     1 |
| 100 | aiven_thanos                                |        |     2 |
| 101 | aiven_transit_gateway_vpc_attachment        |        |     2 |
| 102 | aiven_valkey                                |        |     2 |
| 103 | aiven_valkey_user                           |        |     2 |
+-----+---------------------------------------------+--------+-------+
|     | TOTAL MIGRATED 20%                          | 36     |   180 |
+-----+---------------------------------------------+--------+-------+
```
