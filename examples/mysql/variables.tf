variable "aiven_token" {
  description = "Aiven token"
  type        = string
}

variable "aiven_project_name" {
  description = "Name of an Aiven project assigned to a billing group"
  type        = string
}

variable "mysql__name" {
  description = "Name of the MySQL service"
  type        = string
  default     = "example-mysql-service"
}

variable "mysql_username" {
  description = "MySQL username"
  type        = string
  default     = "admin"
}

variable "mysql_password" {
  description = "MySQL service user password"
  type        = string
}
