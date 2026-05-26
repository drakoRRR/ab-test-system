variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region"
}

variable "cluster_id" {
  type        = string
  description = "Managed Kafka cluster ID"
}

variable "subnet_id" {
  type        = string
  description = "Subnet ID (projects/{project}/regions/{region}/subnetworks/{name}) for cluster network access"
}

variable "vcpu_count" {
  type        = number
  description = "Total vCPU count across the cluster (minimum 3)"
  default     = 3
}

variable "memory_gb" {
  type        = number
  description = "Total memory in GB across the cluster (minimum 3)"
  default     = 3
}

variable "topics" {
  type = map(object({
    partition_count    = number
    replication_factor = number
  }))
  description = "Kafka topics to create"
  default     = {}
}
