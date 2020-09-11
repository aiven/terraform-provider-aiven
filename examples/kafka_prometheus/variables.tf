
variable aiven_api_token {
  type = string
}

variable project {
  type = string
}

variable kafka_svc {
  type    = string
  default = "tf-kafka"
}

variable prom_name {
  type    = string
  default = "Prometheus TF Example"
}
