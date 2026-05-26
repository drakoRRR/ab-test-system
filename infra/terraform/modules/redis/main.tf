resource "google_redis_instance" "this" {
  project            = var.project_id
  name               = var.name
  region             = var.region
  tier               = var.tier
  memory_size_gb     = var.memory_size_gb
  redis_version      = var.redis_version
  authorized_network = var.authorized_network
  auth_enabled       = var.auth_enabled

  transit_encryption_mode = "SERVER_AUTHENTICATION"
}
