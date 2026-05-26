output "email" {
  value       = google_service_account.this.email
  description = "Service account email"
}

output "id" {
  value       = google_service_account.this.id
  description = "Service account fully-qualified resource name"
}
