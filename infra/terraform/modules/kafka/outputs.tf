output "bootstrap_address" {
  value       = google_managed_kafka_cluster.this.bootstrap_address
  description = "Kafka bootstrap endpoint (host:port)"
}

output "cluster_id" {
  value       = google_managed_kafka_cluster.this.cluster_id
  description = "Managed Kafka cluster ID"
}
