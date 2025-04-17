variable "aiven_token" {
  description = "Aiven token"
  type        = string
  sensitive   = true
}

variable "aws_account_id" {
  description = "AWS account ID"
  type        = string
}

variable "aiven_project_name" {
  description = "Name of the Aiven project."
  type        = string
}