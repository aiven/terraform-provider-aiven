variable "aiven_token" {
  description = "Aiven token"
  type        = string
}

variable "aiven_project" {
  description = "Aiven project name"
  type        = string
}

variable "aiven_organization" {
  description = "Aiven organization name"
  type        = string
}

// MongoDB URI format: mongodb://user:password@host:port
variable "mongo_uri" {
  description = "MongoDB URI"
  type        = string
}
