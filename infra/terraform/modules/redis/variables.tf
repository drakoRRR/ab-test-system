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
  description = "Memorystore instance name"
}

variable "authorized_network" {
  type        = string
  description = "VPC network ID authorized to access this instance"
}

variable "tier" {
  type        = string
  description = "BASIC for no replication, STANDARD_HA for read replica"
  default     = "BASIC"

  validation {
    condition     = contains(["BASIC", "STANDARD_HA"], var.tier)
    error_message = "tier must be BASIC or STANDARD_HA"
  }
}

variable "memory_size_gb" {
  type        = number
  description = "Redis memory capacity in GB"
  default     = 1
}

variable "redis_version" {
  type        = string
  description = "Redis engine version"
  default     = "REDIS_7_2"
}

variable "auth_enabled" {
  type        = bool
  description = "Enable AUTH string for password authentication"
  default     = false
}
