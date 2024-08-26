variable "aiven_api_token" {
  description = "Aiven API token"
  type        = string
  sensitive = true
}

# Pre-existing Aiven project
variable "aiven_project_name" {
  description = "Aiven project name"
  type        = string
}


variable "s3_bucket_access_key" {
  default = "AAAAAAAAAAAAAAAAAAA"
  type        = string
}

variable "s3_bucket_secret_key" {
  default = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
  type        = string
  sensitive = true
}

variable "s3_bucket_url" {
  default = "https://mybucket.s3-myregion.amazonaws.com/mydataset/"
  type        = string
}
