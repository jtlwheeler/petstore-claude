output "cloud_run_url" {
  description = "The URL of the deployed Cloud Run service"
  value       = google_cloud_run_v2_service.petstore.uri
}

output "db_instance_connection_name" {
  description = "The connection name of the Cloud SQL instance"
  value       = google_sql_database_instance.petstore.connection_name
}

output "db_private_ip" {
  description = "The private IP address of the Cloud SQL instance"
  value       = google_sql_database_instance.petstore.private_ip_address
}

output "tfstate_bucket_name" {
  description = "The name of the GCS bucket used for Terraform state"
  value       = google_storage_bucket.tfstate.name
}
