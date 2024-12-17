# When both USER and CLIENT_ID are specified
terraform import aiven_kafka_quota.example_quota PROJECT/SERVICE_NAME/CLIENT_ID/USER
# When only USER is specified
terraform import aiven_kafka_quota.example_quota PROJECT/SERVICE_NAME//USER
# When only CLIENT_ID is specified
terraform import aiven_kafka_quota.example_quota PROJECT/SERVICE_NAME/CLIENT_ID/
