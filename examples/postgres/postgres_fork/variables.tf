variable "aiven_token" {
  description = "Aiven token"
  type        = string
}

variable "aiven_project" {
  description = "Aiven project name"
  type        = string
}

variable "source_pg_name" {
  description = "Name of the source PostgreSQL service"
  type        = string
  default     = "example-source-postgres"
}

variable "pg_fork_name" {
  description = "Name of the forked PostgreSQL service"
  type        = string
  default     = "example-forked-postgres"
}