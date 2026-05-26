terraform {
  backend "gcs" {
    bucket = "splitlab-tfstate"
    prefix = "production"
  }
}
