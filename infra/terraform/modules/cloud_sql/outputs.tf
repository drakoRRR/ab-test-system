output "private_ip" {
  value       = google_sql_database_instance.this.private_ip_address
  description = "Private IP address of the Cloud SQL instance"
}

output "instance_name" {
  value       = google_sql_database_instance.this.name
  description = "Cloud SQL instance name"
}

output "connection_name" {
  value       = google_sql_database_instance.this.connection_name
  description = "Instance connection name in project:region:instance format"
}

output "database_name" {
  value       = google_sql_database.this.name
  description = "PostgreSQL database name"
}

output "db_user" {
  value       = google_sql_user.this.name
  description = "PostgreSQL user name"
}
