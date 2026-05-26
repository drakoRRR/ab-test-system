variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "secret_id" {
  type        = string
  description = "Secret identifier within the project"
}

variable "secret_data" {
  type        = string
  description = "Secret value. If null, the secret resource is created but no version is set — value must be populated manually"
  sensitive   = true
  default     = null
}

variable "accessors" {
  type        = list(string)
  description = "IAM members granted roles/secretmanager.secretAccessor"
  default     = []
}
