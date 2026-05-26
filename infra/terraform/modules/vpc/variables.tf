variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region for all resources"
}

variable "network_name" {
  type        = string
  description = "VPC network name"
}

variable "subnet_cidr" {
  type        = string
  description = "Primary CIDR range for the subnet (e.g. 10.0.0.0/24)"
}

variable "private_services_cidr" {
  type        = string
  description = "CIDR range reserved for private service access — Cloud SQL private IP (e.g. 10.1.0.0/16)"
}

variable "connector_name" {
  type        = string
  description = "VPC Access Connector name — used by Cloud Run to reach private resources"
}

variable "connector_cidr" {
  type        = string
  description = "CIDR /28 range for the VPC Access Connector, must not overlap with subnet_cidr (e.g. 10.8.0.0/28)"
}

variable "connector_min_instances" {
  type        = number
  description = "Minimum number of connector instances"
  default     = 2
}

variable "connector_max_instances" {
  type        = number
  description = "Maximum number of connector instances"
  default     = 3
}

variable "connector_machine_type" {
  type        = string
  description = "Machine type for connector instances — e2-micro (staging) or e2-standard-4 (production)"
  default     = "e2-micro"
}
