terraform {
  backend "s3" {
    bucket       = "url-shortener-188735374190"
    key          = "splitit-eks/terraform.tfstate"
    region       = "eu-west-2"
    encrypt      = true
    use_lockfile = true
  }
}
