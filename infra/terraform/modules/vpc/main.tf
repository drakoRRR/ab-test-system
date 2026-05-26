resource "google_compute_network" "this" {
  name                    = var.network_name
  project                 = var.project_id
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "this" {
  name          = "${var.network_name}-subnet"
  project       = var.project_id
  region        = var.region
  network       = google_compute_network.this.id
  ip_cidr_range = var.subnet_cidr
}

# Reserve an IP range for Cloud SQL and other managed services to use
# when establishing a private connection into the VPC.
resource "google_compute_global_address" "private_services" {
  name          = "${var.network_name}-private-services"
  project       = var.project_id
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = split("/", var.private_services_cidr)[1]
  address       = split("/", var.private_services_cidr)[0]
  network       = google_compute_network.this.id
}

resource "google_service_networking_connection" "private_services" {
  network                 = google_compute_network.this.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_services.name]
}

resource "google_vpc_access_connector" "this" {
  name    = var.connector_name
  project = var.project_id
  region  = var.region

  subnet {
    name = google_compute_subnetwork.this.name
  }

  min_instances = var.connector_min_instances
  max_instances = var.connector_max_instances
  machine_type  = var.connector_machine_type
}

# Allow traffic between resources inside the VPC
resource "google_compute_firewall" "internal" {
  name    = "${var.network_name}-allow-internal"
  project = var.project_id
  network = google_compute_network.this.name

  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  allow {
    protocol = "icmp"
  }

  source_ranges = [var.subnet_cidr, var.connector_cidr]
}
