output "secret_id" {
  value       = google_secret_manager_secret.this.secret_id
  description = "Secret ID within the project"
}

output "secret_name" {
  value       = google_secret_manager_secret.this.name
  description = "Fully-qualified secret resource name (projects/{project}/secrets/{id})"
}
