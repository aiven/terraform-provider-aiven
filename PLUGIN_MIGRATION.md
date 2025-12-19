# Plugin Framework Migration Status

```
+----+---------------------------------------------+--------+-------+
|  # | RESOURCE NAME                               | PLUGIN | COUNT |
+----+---------------------------------------------+--------+-------+
|  1 | aiven_account                               |        |     2 |
|  2 | aiven_account_authentication                |        |     2 |
|  3 | aiven_account_team                          |        |     2 |
|  4 | aiven_account_team_member                   |        |     2 |
|  5 | aiven_account_team_project                  |        |     2 |
|  6 | aiven_aws_privatelink                       |        |     2 |
|  7 | aiven_aws_vpc_peering_connection            |        |     2 |
|  8 | aiven_azure_privatelink                     |        |     2 |
|  9 | aiven_azure_privatelink_connection_approval |        |     1 |
| 10 | aiven_azure_vpc_peering_connection          |        |     2 |
| 11 | aiven_billing_group                         | yes    |     2 |
| 12 | aiven_cassandra                             |        |     2 |
| 13 | aiven_cassandra_user                        |        |     2 |
| 14 | aiven_clickhouse                            |        |     2 |
| 15 | aiven_clickhouse_database                   |        |     2 |
| 16 | aiven_clickhouse_grant                      |        |     1 |
| 17 | aiven_clickhouse_role                       |        |     1 |
| 18 | aiven_clickhouse_user                       |        |     2 |
| 19 | aiven_connection_pool                       |        |     2 |
| 20 | aiven_dragonfly                             |        |     2 |
| 21 | aiven_flink                                 |        |     2 |
| 22 | aiven_flink_application                     |        |     2 |
| 23 | aiven_flink_application_deployment          |        |     1 |
| 24 | aiven_flink_application_version             |        |     2 |
| 25 | aiven_gcp_privatelink                       |        |     2 |
| 26 | aiven_gcp_privatelink_connection_approval   |        |     1 |
| 27 | aiven_gcp_vpc_peering_connection            |        |     2 |
| 28 | aiven_governance_access                     | yes    |     1 |
| 29 | aiven_grafana                               |        |     2 |
| 30 | aiven_influxdb                              |        |     2 |
| 31 | aiven_influxdb_database                     |        |     2 |
| 32 | aiven_influxdb_user                         |        |     2 |
| 33 | aiven_kafka                                 |        |     2 |
| 34 | aiven_kafka_acl                             |        |     2 |
| 35 | aiven_kafka_connect                         |        |     2 |
| 36 | aiven_kafka_connector                       |        |     2 |
| 37 | aiven_kafka_mirrormaker                     |        |     2 |
| 38 | aiven_kafka_native_acl                      |        |     1 |
| 39 | aiven_kafka_quota                           |        |     1 |
| 40 | aiven_kafka_schema                          |        |     2 |
| 41 | aiven_kafka_schema_configuration            |        |     2 |
| 42 | aiven_kafka_schema_registry_acl             |        |     2 |
| 43 | aiven_kafka_topic                           |        |     2 |
| 44 | aiven_kafka_user                            |        |     2 |
| 45 | aiven_m3aggregator                          |        |     2 |
| 46 | aiven_m3db                                  |        |     2 |
| 47 | aiven_m3db_user                             |        |     2 |
| 48 | aiven_mirrormaker_replication_flow          |        |     2 |
| 49 | aiven_mysql                                 |        |     2 |
| 50 | aiven_mysql_database                        | yes    |     2 |
| 51 | aiven_mysql_user                            |        |     2 |
| 52 | aiven_opensearch                            |        |     2 |
| 53 | aiven_opensearch_acl_config                 |        |     2 |
| 54 | aiven_opensearch_acl_rule                   |        |     2 |
| 55 | aiven_opensearch_security_plugin_config     |        |     2 |
| 56 | aiven_opensearch_user                       |        |     2 |
| 57 | aiven_organization                          | yes    |     2 |
| 58 | aiven_organization_address                  | yes    |     2 |
| 59 | aiven_organization_application_user         | yes    |     2 |
| 60 | aiven_organization_application_user_token   | yes    |     1 |
| 61 | aiven_organization_billing_group            | yes    |     2 |
| 62 | aiven_organization_billing_group_list       | yes    |     1 |
| 63 | aiven_organization_group_project            | yes    |     1 |
| 64 | aiven_organization_permission               | yes    |     1 |
| 65 | aiven_organization_project                  | yes    |     2 |
| 66 | aiven_organization_user                     |        |     2 |
| 67 | aiven_organization_user_group               |        |     2 |
| 68 | aiven_organization_user_group_list          | yes    |     1 |
| 69 | aiven_organization_user_group_member        | yes    |     1 |
| 70 | aiven_organization_user_group_member_list   | yes    |     1 |
| 71 | aiven_organization_user_list                | yes    |     1 |
| 72 | aiven_organizational_unit                   | yes    |     2 |
| 73 | aiven_pg                                    |        |     2 |
| 74 | aiven_pg_database                           |        |     2 |
| 75 | aiven_pg_user                               |        |     2 |
| 76 | aiven_project                               |        |     2 |
| 77 | aiven_project_user                          |        |     2 |
| 78 | aiven_project_vpc                           |        |     2 |
| 79 | aiven_redis                                 |        |     2 |
| 80 | aiven_redis_user                            |        |     2 |
| 81 | aiven_service_component                     |        |     1 |
| 82 | aiven_service_integration                   |        |     2 |
| 83 | aiven_service_integration_endpoint          |        |     2 |
| 84 | aiven_service_plan                          | yes    |     1 |
| 85 | aiven_service_plan_list                     | yes    |     1 |
| 86 | aiven_static_ip                             |        |     1 |
| 87 | aiven_thanos                                |        |     2 |
| 88 | aiven_transit_gateway_vpc_attachment        |        |     2 |
| 89 | aiven_valkey                                |        |     2 |
| 90 | aiven_valkey_user                           |        |     2 |
+----+---------------------------------------------+--------+-------+
|    | TOTAL MIGRATED 17%                          | 27     |   160 |
+----+---------------------------------------------+--------+-------+
```
