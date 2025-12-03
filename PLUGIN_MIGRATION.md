# Plugin Framework Migration Status

This document tracks the migration status of all resources and datasources from the old SDK to the new Plugin framework.

## Summary as of 2025-12-01

- Unique items: 100 (table rows)
- Plugin implementations: 18
- Remaining: 82 SDK implementations to be migrated, 19 of which are deprecated

## Resources and Datasources

| #   | Name                                          | Type | Plugin | Deprecated |
|-----|-----------------------------------------------|------|--------|------------|
| 1   | aiven_account                                 | RD   |        | ✓          |
| 2   | aiven_account_authentication                  | RD   |        | ✓          |
| 3   | aiven_account_team                            | RD   |        | ✓          |
| 4   | aiven_account_team_member                     | RD   |        | ✓          |
| 5   | aiven_account_team_project                    | RD   |        | ✓          |
| 6   | aiven_alloydbomni                             | RD   |        | ✓          |
| 7   | aiven_alloydbomni_database                    | RD   |        | ✓          |
| 8   | aiven_alloydbomni_user                        | RD   |        | ✓          |
| 9   | aiven_aws_org_vpc_peering_connection          | RD   |        |            |
| 10  | aiven_aws_privatelink                         | RD   |        |            |
| 11  | aiven_aws_vpc_peering_connection              | RD   |        |            |
| 12  | aiven_azure_org_vpc_peering_connection        | RD   |        |            |
| 13  | aiven_azure_privatelink                       | RD   |        |            |
| 14  | aiven_azure_privatelink_connection_approval   | R    |        |            |
| 15  | aiven_azure_vpc_peering_connection            | RD   |        |            |
| 16  | aiven_billing_group                           | RD   | ✓      |            |
| 17  | aiven_cassandra                               | RD   |        | ✓          |
| 18  | aiven_cassandra_user                          | RD   |        | ✓          |
| 19  | aiven_clickhouse                              | RD   |        |            |
| 20  | aiven_clickhouse_database                     | RD   |        |            |
| 21  | aiven_clickhouse_grant                        | R    |        |            |
| 22  | aiven_clickhouse_role                         | R    |        |            |
| 23  | aiven_clickhouse_user                         | RD   |        |            |
| 24  | aiven_connection_pool                         | RD   |        |            |
| 25  | aiven_dragonfly                               | RD   |        |            |
| 26  | aiven_external_identity                       | D    | ✓      |            |
| 27  | aiven_flink                                   | RD   |        |            |
| 28  | aiven_flink_application                       | RD   |        |            |
| 29  | aiven_flink_application_deployment            | R    |        |            |
| 30  | aiven_flink_application_version               | RD   |        |            |
| 31  | aiven_flink_jar_application                   | R    |        |            |
| 32  | aiven_flink_jar_application_deployment        | R    |        |            |
| 33  | aiven_flink_jar_application_version           | R    |        |            |
| 34  | aiven_gcp_org_vpc_peering_connection          | RD   |        |            |
| 35  | aiven_gcp_privatelink                         | RD   |        |            |
| 36  | aiven_gcp_privatelink_connection_approval     | R    |        |            |
| 37  | aiven_gcp_vpc_peering_connection              | RD   |        |            |
| 38  | aiven_governance_access                       | R    | ✓      |            |
| 39  | aiven_grafana                                 | RD   |        |            |
| 40  | aiven_influxdb                                | RD   |        | ✓          |
| 41  | aiven_influxdb_database                       | RD   |        | ✓          |
| 42  | aiven_influxdb_user                           | RD   |        | ✓          |
| 43  | aiven_kafka                                   | RD   |        |            |
| 44  | aiven_kafka_acl                               | RD   |        |            |
| 45  | aiven_kafka_connect                           | RD   |        |            |
| 46  | aiven_kafka_connector                         | RD   |        |            |
| 47  | aiven_kafka_mirrormaker                       | RD   |        |            |
| 48  | aiven_kafka_native_acl                        | R    |        |            |
| 49  | aiven_kafka_quota                             | R    |        |            |
| 50  | aiven_kafka_schema                            | RD   |        |            |
| 51  | aiven_kafka_schema_configuration              | RD   |        |            |
| 52  | aiven_kafka_schema_registry_acl               | RD   |        |            |
| 53  | aiven_kafka_topic                             | RD   |        |            |
| 54  | aiven_kafka_user                              | RD   |        |            |
| 55  | aiven_m3aggregator                            | RD   |        |            |
| 56  | aiven_m3db                                    | RD   |        | ✓          |
| 57  | aiven_m3db_user                               | RD   |        | ✓          |
| 58  | aiven_mirrormaker_replication_flow            | RD   |        |            |
| 59  | aiven_mysql                                   | RD   |        |            |
| 60  | aiven_mysql_database                          | RD   |        |            |
| 61  | aiven_mysql_user                              | RD   |        |            |
| 62  | aiven_opensearch                              | RD   |        |            |
| 63  | aiven_opensearch_acl_config                   | RD   |        |            |
| 64  | aiven_opensearch_acl_rule                     | RD   |        |            |
| 65  | aiven_opensearch_security_plugin_config       | RD   |        |            |
| 66  | aiven_opensearch_user                         | RD   |        |            |
| 67  | aiven_organization                            | RD   | ✓      |            |
| 68  | aiven_organization_address                    | RD   | ✓      |            |
| 69  | aiven_organization_application_user           | RD   | ✓      |            |
| 70  | aiven_organization_application_user_token     | R    | ✓      |            |
| 71  | aiven_organization_billing_group              | RD   | ✓      |            |
| 72  | aiven_organization_billing_group_list         | D    | ✓      |            |
| 73  | aiven_organization_group_project              | R    | ✓      |            |
| 74  | aiven_organization_permission                 | R    | ✓      |            |
| 75  | aiven_organization_project                    | RD   | ✓      |            |
| 76  | aiven_organization_user                       | RD   |        | ✓          |
| 77  | aiven_organization_user_group                 | RD   |        |            |
| 78  | aiven_organization_user_group_list            | D    | ✓      |            |
| 79  | aiven_organization_user_group_member          | R    | ✓      |            |
| 80  | aiven_organization_user_list                  | D    | ✓      |            |
| 81  | aiven_organization_vpc                        | RD   |        |            |
| 82  | aiven_organizational_unit                     | RD   |        |            |
| 83  | aiven_pg                                      | RD   |        |            |
| 84  | aiven_pg_database                             | RD   |        |            |
| 85  | aiven_pg_user                                 | RD   |        |            |
| 86  | aiven_project                                 | RD   |        |            |
| 87  | aiven_project_user                            | RD   |        | ✓          |
| 88  | aiven_project_vpc                             | RD   |        |            |
| 89  | aiven_redis                                   | RD   |        | ✓          |
| 90  | aiven_redis_user                              | RD   |        | ✓          |
| 91  | aiven_service_component                       | D    |        |            |
| 92  | aiven_service_integration                     | RD   |        |            |
| 93  | aiven_service_integration_endpoint            | RD   |        |            |
| 94  | aiven_service_plan                            | D    | ✓      |            |
| 95  | aiven_service_plan_list                       | D    | ✓      |            |
| 96  | aiven_static_ip                               | R    |        |            |
| 97  | aiven_thanos                                  | RD   |        |            |
| 98  | aiven_transit_gateway_vpc_attachment          | RD   |        |            |
| 99  | aiven_valkey                                  | RD   |        |            |
| 100 | aiven_valkey_user                             | RD   |        |            |
