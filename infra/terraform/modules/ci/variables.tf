variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "pool_id" {
  type        = string
  description = "Workload Identity Pool ID"
}

variable "github_repo" {
  type        = string
  description = "GitHub repository in owner/name format (e.g. acme/my-app)"
}

variable "deploy_sa_email" {
  type        = string
  description = "Service account email that GitHub Actions will impersonate"
}
