output "lb_ip" {
  value       = module.load_balancer.ip_address
  description = "Static external IP of the load balancer — point your DNS A record here"
}

output "server_url" {
  value       = module.server.url
  description = "Direct Cloud Run URL (internal — use lb_ip for public access)"
}

output "registry_url" {
  value       = module.artifact_registry.url
  description = "Docker registry base URL"
}

output "wif_provider" {
  value       = module.ci.workload_identity_provider
  description = "Workload Identity provider resource name — set as WIF_PROVIDER in GitHub Secrets"
}

output "deploy_sa_email" {
  value       = module.deploy_sa.email
  description = "Deploy service account email — set as DEPLOY_SA in GitHub Secrets"
}
