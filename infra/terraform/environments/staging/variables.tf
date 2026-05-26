variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region for all resources"
  default     = "us-central1"
}

variable "github_repo" {
  type        = string
  description = "GitHub repository in owner/name format — used for Workload Identity binding"
}

variable "db_password" {
  type        = string
  description = "PostgreSQL database password"
  sensitive   = true
}

variable "firebase_credentials" {
  type        = string
  description = "Firebase service account credentials JSON"
  sensitive   = true
}
