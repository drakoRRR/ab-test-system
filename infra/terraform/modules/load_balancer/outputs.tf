output "ip_address" {
  value       = google_compute_global_address.this.address
  description = "Static external IP address of the load balancer"
}

output "ssl_certificate_name" {
  value       = length(var.domains) > 0 ? google_compute_managed_ssl_certificate.this[0].name : ""
  description = "Google-managed SSL certificate name (empty when no domains configured)"
}
