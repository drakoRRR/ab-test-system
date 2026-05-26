output "repository_id" {
  value       = google_artifact_registry_repository.this.repository_id
  description = "Artifact Registry repository ID"
}

output "url" {
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.this.repository_id}"
  description = "Base URL for Docker images in this repository"
}
