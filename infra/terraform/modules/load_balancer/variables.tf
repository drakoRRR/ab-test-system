variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region where the Cloud Run service lives"
}

variable "name" {
  type        = string
  description = "Name prefix for all load balancer resources"
}

variable "cloud_run_service_name" {
  type        = string
  description = "Cloud Run service name to route traffic to"
}

variable "domains" {
  type        = list(string)
  description = "Domains for Google-managed SSL certificate. Empty list skips cert creation (no custom domain)."
  default     = []
}
