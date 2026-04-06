resource "random_password" "db" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# GCS bucket for Terraform state
resource "google_storage_bucket" "tfstate" {
  name          = var.tfstate_bucket_name
  location      = var.region
  force_destroy = false

  versioning {
    enabled = true
  }

  uniform_bucket_level_access = true
}

# VPC network for private Cloud SQL connectivity
resource "google_compute_network" "petstore" {
  name                    = "petstore-vpc-${var.environment}"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "petstore" {
  name          = "petstore-subnet-${var.environment}"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.petstore.id
}

resource "google_compute_global_address" "private_ip_range" {
  name          = "petstore-private-ip-${var.environment}"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.petstore.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.petstore.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_range.name]
}

# VPC connector for Cloud Run to access private network
resource "google_vpc_access_connector" "petstore" {
  name          = "petstore-connector-${var.environment}"
  region        = var.region
  ip_cidr_range = "10.8.0.0/28"
  network       = google_compute_network.petstore.name
}

# Cloud SQL PostgreSQL instance
resource "google_sql_database_instance" "petstore" {
  name             = var.db_instance_name
  database_version = "POSTGRES_17"
  region           = var.region

  deletion_protection = false

  settings {
    tier = "db-f1-micro"

    backup_configuration {
      enabled = true
    }

    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.petstore.id
    }
  }

  depends_on = [google_service_networking_connection.private_vpc_connection]
}

resource "google_sql_database" "petstore" {
  name     = var.db_name
  instance = google_sql_database_instance.petstore.name
}

resource "google_sql_user" "petstore" {
  name     = var.db_user
  instance = google_sql_database_instance.petstore.name
  password = random_password.db.result
}

# Cloud Run service
resource "google_cloud_run_v2_service" "petstore" {
  name     = "petstore-${var.environment}"
  location = var.region

  template {
    containers {
      image = var.container_image

      env {
        name  = "DB_HOST"
        value = google_sql_database_instance.petstore.private_ip_address
      }

      env {
        name  = "DB_NAME"
        value = var.db_name
      }

      env {
        name  = "DB_USER"
        value = var.db_user
      }

      env {
        name  = "DB_PASSWORD"
        value = random_password.db.result
      }

      env {
        name  = "DB_PORT"
        value = "5432"
      }
    }

    vpc_access {
      connector = google_vpc_access_connector.petstore.id
      egress    = "PRIVATE_RANGES_ONLY"
    }
  }
}

# Allow unauthenticated access to Cloud Run
resource "google_cloud_run_v2_service_iam_member" "public" {
  location = google_cloud_run_v2_service.petstore.location
  name     = google_cloud_run_v2_service.petstore.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
