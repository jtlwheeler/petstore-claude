terraform {
  backend "gcs" {
    bucket = "petstore-tfstate-dev"
    prefix = "terraform/state"
  }
}
