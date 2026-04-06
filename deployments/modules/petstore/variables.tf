variable "region" {
  description = "The GCP region to deploy resources"
  type        = string
}

variable "environment" {
  description = "The deployment environment"
  type        = string
}

variable "tfstate_bucket_name" {
  description = "The name of the GCS bucket used for Terraform state storage"
  type        = string
}

variable "container_image" {
  description = "The container image to deploy to Cloud Run"
  type        = string
}

variable "db_instance_name" {
  description = "The name of the Cloud SQL instance"
  type        = string
}

variable "db_name" {
  description = "The name of the PostgreSQL database"
  type        = string
}

variable "db_user" {
  description = "The PostgreSQL database user"
  type        = string
}

