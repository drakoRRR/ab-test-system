output "workload_identity_provider" {
  value       = google_iam_workload_identity_pool_provider.github.name
  description = "Full provider resource name — use as workload_identity_provider in google-github-actions/auth"
}
