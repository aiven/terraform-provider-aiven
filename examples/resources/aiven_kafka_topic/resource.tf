resource "aiven_kafka_topic" "mytesttopic" {
    project = aiven_project.myproject.project
    service_name = aiven_kafka.myservice.service_name
    topic_name = "<TOPIC_NAME>"
    partitions = 5
    replication = 3
    termination_protection = true
    
    config {
        flush_ms = 10
        unclean_leader_election_enable = true
        cleanup_policy = "compact,delete"
    }


    timeouts {
        create = "1m"
        read = "5m"
    }
}
