module "vpc" {
  source       = "../../modules/vpc"
  project_id   = var.project_id
  region       = var.region
  network_name = "splitlab-production"

  subnet_cidr           = "10.2.0.0/24"
  private_services_cidr = "10.3.0.0/16"

  connector_name          = "splitlab-production-connector"
  connector_cidr          = "10.9.0.0/28"
  connector_min_instances = 3
  connector_max_instances = 7
  connector_machine_type  = "e2-standard-4"
}

module "redis" {
  source     = "../../modules/redis"
  project_id = var.project_id
  region     = var.region
  name       = "splitlab-production"

  authorized_network = module.vpc.network_id
  tier               = "STANDARD_HA"
  memory_size_gb     = 4
  auth_enabled       = true

  depends_on = [module.vpc]
}

locals {
  secrets = {
    "db-password" = {
      data      = var.db_password
      accessors = [module.server_sa.email, module.consumer_sa.email]
    }
    "firebase-credentials" = {
      data      = var.firebase_credentials
      accessors = [module.server_sa.email]
    }
    "redis-auth-string" = {
      data      = module.redis.auth_string
      accessors = [module.server_sa.email, module.consumer_sa.email]
    }
  }
}

module "secrets" {
  for_each = local.secrets

  source      = "../../modules/secret_manager"
  project_id  = var.project_id
  secret_id   = each.key
  secret_data = each.value.data
  accessors   = each.value.accessors
}

module "kafka" {
  source     = "../../modules/kafka"
  project_id = var.project_id
  region     = var.region
  cluster_id = "splitlab-production"
  subnet_id  = module.vpc.subnet_id
  vcpu_count = 6
  memory_gb  = 6

  topics = {
    "ab-test-events" = {
      partition_count    = 6
      replication_factor = 3
    }
  }

  depends_on = [module.vpc]
}

module "cloud_sql" {
  source     = "../../modules/cloud_sql"
  project_id = var.project_id
  region     = var.region
  network_id = module.vpc.network_id

  instance_name     = "splitlab-production"
  database_name     = "splitlab"
  db_user           = "splitlab"
  db_password       = var.db_password
  tier              = "db-g1-small"
  disk_size_gb      = 50
  availability_type = "REGIONAL"
  deletion_protection = true

  depends_on = [module.vpc]
}

module "artifact_registry" {
  source        = "../../modules/artifact_registry"
  project_id    = var.project_id
  region        = var.region
  repository_id = "splitlab"
  description   = "SplitLab Docker images (production)"
  writers       = [module.deploy_sa.email]
  readers       = [module.server_sa.email, module.consumer_sa.email]
}

module "server" {
  source     = "../../modules/cloud_run"
  project_id = var.project_id
  region     = var.region
  name       = "splitlab-server"
  image      = "${var.region}-docker.pkg.dev/${var.project_id}/splitlab/server:latest"

  service_account_email = module.server_sa.email
  vpc_connector_id      = module.vpc.connector_id
  allow_unauthenticated = true
  ingress               = "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"
  cpu_idle              = true
  min_instances         = 1
  max_instances         = 20
  cpu                   = "2"
  memory                = "1Gi"

  env_vars = {
    DB_HOST      = module.cloud_sql.private_ip
    DB_NAME      = module.cloud_sql.database_name
    DB_USER      = module.cloud_sql.db_user
    REDIS_HOST   = module.redis.host
    REDIS_PORT   = tostring(module.redis.port)
    KAFKA_BROKER = module.kafka.bootstrap_address
    ENVIRONMENT  = "production"
  }

  secret_env_vars = {
    DB_PASSWORD          = { secret = "db-password" }
    FIREBASE_CREDENTIALS = { secret = "firebase-credentials" }
    REDIS_PASSWORD       = { secret = "redis-auth-string" }
  }

  depends_on = [module.vpc, module.secrets]
}

module "consumer" {
  source     = "../../modules/cloud_run"
  project_id = var.project_id
  region     = var.region
  name       = "splitlab-consumer"
  image      = "${var.region}-docker.pkg.dev/${var.project_id}/splitlab/consumer:latest"

  service_account_email = module.consumer_sa.email
  vpc_connector_id      = module.vpc.connector_id
  allow_unauthenticated = false
  ingress               = "INGRESS_TRAFFIC_INTERNAL_ONLY"
  cpu_idle              = false
  min_instances         = 2
  max_instances         = 10
  cpu                   = "2"
  memory                = "1Gi"

  env_vars = {
    DB_HOST      = module.cloud_sql.private_ip
    DB_NAME      = module.cloud_sql.database_name
    DB_USER      = module.cloud_sql.db_user
    REDIS_HOST   = module.redis.host
    REDIS_PORT   = tostring(module.redis.port)
    KAFKA_BROKER = module.kafka.bootstrap_address
    ENVIRONMENT  = "production"
  }

  secret_env_vars = {
    DB_PASSWORD    = { secret = "db-password" }
    REDIS_PASSWORD = { secret = "redis-auth-string" }
  }

  depends_on = [module.vpc, module.secrets]
}

module "server_sa" {
  source       = "../../modules/iam"
  project_id   = var.project_id
  account_id   = "splitlab-server"
  display_name = "SplitLab API Server"
  roles = [
    "roles/secretmanager.secretAccessor",
    "roles/cloudsql.client",
    "roles/managedkafka.producer",
  ]
}

module "consumer_sa" {
  source       = "../../modules/iam"
  project_id   = var.project_id
  account_id   = "splitlab-consumer"
  display_name = "SplitLab Kafka Consumer"
  roles = [
    "roles/secretmanager.secretAccessor",
    "roles/cloudsql.client",
    "roles/managedkafka.consumer",
  ]
}

module "load_balancer" {
  source     = "../../modules/load_balancer"
  project_id = var.project_id
  region     = var.region
  name       = "splitlab-production"

  cloud_run_service_name = module.server.service_name
  domains                = [var.api_domain]
}

module "ci" {
  source     = "../../modules/ci"
  project_id = var.project_id
  pool_id    = "splitlab-production"

  github_repo     = var.github_repo
  deploy_sa_email = module.deploy_sa.email
}

module "deploy_sa" {
  source       = "../../modules/iam"
  project_id   = var.project_id
  account_id   = "splitlab-deploy"
  display_name = "SplitLab GitHub Actions Deploy"
  roles = [
    "roles/run.developer",
    "roles/artifactregistry.writer",
    "roles/iam.serviceAccountUser",
  ]
}
