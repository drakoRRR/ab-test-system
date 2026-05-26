variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region"
}

variable "repository_id" {
  type        = string
  description = "Artifact Registry repository ID"
}

variable "description" {
  type        = string
  description = "Human-readable description of the repository"
  default     = ""
}

variable "writers" {
  type        = list(string)
  description = "List of IAM members granted roles/artifactregistry.writer (e.g. serviceAccount:...)"
  default     = []
}

variable "readers" {
  type        = list(string)
  description = "List of IAM members granted roles/artifactregistry.reader (e.g. serviceAccount:...)"
  default     = []
}
