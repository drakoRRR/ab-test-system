output "network_id" {
  value       = google_compute_network.this.id
  description = "VPC network self-link ID"
}

output "network_name" {
  value       = google_compute_network.this.name
  description = "VPC network name"
}

output "subnet_name" {
  value       = google_compute_subnetwork.this.name
  description = "Subnet name"
}

output "subnet_id" {
  value       = google_compute_subnetwork.this.id
  description = "Subnet self-link ID"
}

output "connector_id" {
  value       = google_vpc_access_connector.this.id
  description = "VPC Access Connector self-link ID"
}

output "private_services_connection" {
  value       = google_service_networking_connection.private_services.id
  description = "Private services peering connection ID"
}
