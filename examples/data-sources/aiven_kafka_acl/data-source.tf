data "aiven_kafka_acl" "mytestacl" {
  project      = aiven_project.myproject.project
  service_name = aiven_kafka.mykafka.service_name
  topic        = "<TOPIC_NAME_PATTERN>"
  permission   = "<PERMISSON>"
  username     = "<USERNAME_PATTERN>"
}

