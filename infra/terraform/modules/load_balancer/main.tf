resource "google_compute_global_address" "this" {
  project = var.project_id
  name    = var.name
}

# Serverless NEG — bridges the global LB to the Cloud Run service
resource "google_compute_region_network_endpoint_group" "this" {
  project               = var.project_id
  name                  = "${var.name}-neg"
  network_endpoint_type = "SERVERLESS"
  region                = var.region

  cloud_run {
    service = var.cloud_run_service_name
  }
}

resource "google_compute_backend_service" "this" {
  project     = var.project_id
  name        = "${var.name}-backend"
  protocol    = "HTTPS"
  timeout_sec = 30

  backend {
    group = google_compute_region_network_endpoint_group.this.id
  }
}

resource "google_compute_url_map" "this" {
  project         = var.project_id
  name            = var.name
  default_service = google_compute_backend_service.this.id
}

# HTTP → HTTPS redirect map
resource "google_compute_url_map" "redirect" {
  project = var.project_id
  name    = "${var.name}-redirect"

  default_url_redirect {
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
    strip_query            = false
  }
}

resource "google_compute_managed_ssl_certificate" "this" {
  count   = length(var.domains) > 0 ? 1 : 0
  project = var.project_id
  name    = "${var.name}-cert"

  managed {
    domains = var.domains
  }
}

resource "google_compute_target_https_proxy" "this" {
  project = var.project_id
  name    = "${var.name}-https"
  url_map = google_compute_url_map.this.id

  ssl_certificates = length(var.domains) > 0 ? [
    google_compute_managed_ssl_certificate.this[0].id
  ] : []
}

resource "google_compute_target_http_proxy" "redirect" {
  project = var.project_id
  name    = "${var.name}-http"
  url_map = google_compute_url_map.redirect.id
}

resource "google_compute_global_forwarding_rule" "https" {
  project               = var.project_id
  name                  = "${var.name}-https"
  target                = google_compute_target_https_proxy.this.id
  ip_address            = google_compute_global_address.this.address
  port_range            = "443"
  load_balancing_scheme = "EXTERNAL_MANAGED"
}

resource "google_compute_global_forwarding_rule" "http" {
  project               = var.project_id
  name                  = "${var.name}-http"
  target                = google_compute_target_http_proxy.redirect.id
  ip_address            = google_compute_global_address.this.address
  port_range            = "80"
  load_balancing_scheme = "EXTERNAL_MANAGED"
}
