variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region"
}

variable "network_id" {
  type        = string
  description = "VPC network ID used for private IP allocation"
}

variable "instance_name" {
  type        = string
  description = "Cloud SQL instance name"
}

variable "database_version" {
  type        = string
  description = "PostgreSQL engine version"
  default     = "POSTGRES_15"
}

variable "tier" {
  type        = string
  description = "Cloud SQL machine type (e.g. db-f1-micro, db-g1-small)"
  default     = "db-f1-micro"
}

variable "disk_size_gb" {
  type        = number
  description = "Initial disk size in GB"
  default     = 10
}

variable "availability_type" {
  type        = string
  description = "ZONAL for single-AZ, REGIONAL for HA standby"
  default     = "ZONAL"

  validation {
    condition     = contains(["ZONAL", "REGIONAL"], var.availability_type)
    error_message = "availability_type must be ZONAL or REGIONAL"
  }
}

variable "database_name" {
  type        = string
  description = "Name of the PostgreSQL database to create"
}

variable "db_user" {
  type        = string
  description = "PostgreSQL user name"
}

variable "db_password" {
  type        = string
  description = "PostgreSQL user password"
  sensitive   = true
}

variable "deletion_protection" {
  type        = bool
  description = "Prevent Terraform from destroying the instance"
  default     = true
}
