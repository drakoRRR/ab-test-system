resource "google_managed_kafka_cluster" "this" {
  project    = var.project_id
  cluster_id = var.cluster_id
  location   = var.region

  capacity_config {
    vcpu_count   = var.vcpu_count
    memory_bytes = var.memory_gb * 1024 * 1024 * 1024
  }

  gcp_config {
    access_config {
      network_configs {
        subnet = var.subnet_id
      }
    }
  }
}

resource "google_managed_kafka_topic" "topics" {
  for_each = var.topics

  project            = var.project_id
  cluster            = google_managed_kafka_cluster.this.cluster_id
  location           = var.region
  topic_id           = each.key
  partition_count    = each.value.partition_count
  replication_factor = each.value.replication_factor
}
