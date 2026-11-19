variable "project_name" {
  type        = string
  description = "Project name, used as a prefix for IAM role names."
}

variable "oidc_provider_arn" {
  type        = string
  description = "ARN of the EKS cluster's IAM OIDC provider."
}

variable "hosted_zone_arn" {
  type        = string
  description = "ARN of the Route 53 hosted zone ExternalDNS and cert-manager may change."
}

variable "db_password_secret_arn" {
  type        = string
  description = "ARN of the RDS-managed secret External Secrets may read."
}

variable "db_host_parameter_arn" {
  type        = string
  description = "ARN of the SSM parameter External Secrets may read."
}
