data "aiven_aws_privatelink" "foo" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.bar.service_name
}

