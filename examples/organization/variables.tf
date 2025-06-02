variable "aiven_token" {
  description = "Aiven token"
  type        = string
}

variable "organization_name" {
  description = "Name of the Aiven organization"
  type        = string
}

variable "prod_project_name" {
  description = "Prefix of the projects for production environments"
  type        = string
}

variable "qa_project_name" {
  description = "Prefix of the projects for QA environments"
  type        = string
}

variable "dev_project_name" {
  description = "Prefix of the projects for development environments"
  type        = string
}
