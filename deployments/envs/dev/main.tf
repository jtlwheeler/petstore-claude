module "petstore" {
  source = "../../modules/petstore"

  region              = "us-central1"
  environment         = "dev"
  tfstate_bucket_name = "petstore-tfstate-dev"
  container_image     = "gcr.io/your-gcp-project-id/petstore:latest"
  db_instance_name    = "petstore-db-dev"
  db_name             = "petstore"
  db_user             = "petstore"
}
