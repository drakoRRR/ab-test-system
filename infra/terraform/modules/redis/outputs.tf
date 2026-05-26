output "host" {
  value       = google_redis_instance.this.host
  description = "Private IP address of the Redis instance"
}

output "port" {
  value       = google_redis_instance.this.port
  description = "Redis port"
}

output "auth_string" {
  value       = google_redis_instance.this.auth_string
  description = "AUTH password for the Redis instance (empty when auth_enabled = false)"
  sensitive   = true
}
