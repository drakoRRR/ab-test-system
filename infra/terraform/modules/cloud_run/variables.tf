variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region"
}

variable "name" {
  type        = string
  description = "Cloud Run service name"
}

variable "image" {
  type        = string
  description = "Docker image URL (registry/repo/image:tag)"
}

variable "service_account_email" {
  type        = string
  description = "Service account email the revision runs as"
}

variable "vpc_connector_id" {
  type        = string
  description = "VPC Access Connector ID for private network egress"
}

variable "env_vars" {
  type        = map(string)
  description = "Plain environment variables"
  default     = {}
}

variable "secret_env_vars" {
  type = map(object({
    secret  = string
    version = optional(string, "latest")
  }))
  description = "Environment variables sourced from Secret Manager (key = env var name)"
  default     = {}
}

variable "cpu" {
  type        = string
  description = "CPU limit per instance (e.g. '1', '2')"
  default     = "1"
}

variable "memory" {
  type        = string
  description = "Memory limit per instance (e.g. '512Mi', '1Gi')"
  default     = "512Mi"
}

variable "concurrency" {
  type        = number
  description = "Maximum concurrent requests per instance"
  default     = 80
}

variable "min_instances" {
  type        = number
  description = "Minimum number of running instances"
  default     = 0
}

variable "max_instances" {
  type        = number
  description = "Maximum number of running instances"
  default     = 10
}

variable "cpu_idle" {
  type        = bool
  description = "Throttle CPU when not processing requests. Set false for background workers that poll continuously."
  default     = true
}

variable "ingress" {
  type        = string
  description = "Ingress traffic sources: INGRESS_TRAFFIC_ALL, INGRESS_TRAFFIC_INTERNAL_ONLY, INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"
  default     = "INGRESS_TRAFFIC_ALL"

  validation {
    condition = contains([
      "INGRESS_TRAFFIC_ALL",
      "INGRESS_TRAFFIC_INTERNAL_ONLY",
      "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER",
    ], var.ingress)
    error_message = "ingress must be one of: INGRESS_TRAFFIC_ALL, INGRESS_TRAFFIC_INTERNAL_ONLY, INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"
  }
}

variable "allow_unauthenticated" {
  type        = bool
  description = "Bind roles/run.invoker to allUsers (public HTTP endpoint)"
  default     = false
}
