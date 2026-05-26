variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "account_id" {
  type        = string
  description = "Service account ID — unique within the project, used in the email address"
}

variable "display_name" {
  type        = string
  description = "Human-readable name shown in GCP console"
  default     = ""
}

variable "roles" {
  type        = list(string)
  description = "Project-level IAM roles granted to this service account"
  default     = []
}
