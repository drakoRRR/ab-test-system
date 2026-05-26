output "url" {
  value       = google_cloud_run_v2_service.this.uri
  description = "HTTPS URL of the Cloud Run service"
}

output "service_name" {
  value       = google_cloud_run_v2_service.this.name
  description = "Cloud Run service name"
}
