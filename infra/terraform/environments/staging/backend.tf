terraform {
  backend "gcs" {
    bucket = "splitlab-tfstate"
    prefix = "staging"
  }
}
