# General

variable "region" {
  type        = string
  description = "AWS region to deploy resources in."
  default     = "eu-west-2"
}

variable "project_name" {
  type        = string
  description = "Project name, used as a prefix for resource names and in tags."
  default     = "splitit"
}

variable "environment" {
  type        = string
  description = "Environment name used in tags (e.g., prod)."
  default     = "prod"
}

variable "owner" {
  type        = string
  description = "Owner of the resources, used in tags."
  default     = "hamzah"
}

# VPC

variable "vpc_cidr" {
  type        = string
  description = "CIDR block for the VPC."
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  type        = map(string)
  description = "CIDR blocks for the public subnets (NLB, NAT gateway), keyed by availability zone."
  default = {
    "eu-west-2a" = "10.0.1.0/24"
    "eu-west-2b" = "10.0.2.0/24"
  }
}

variable "app_subnet_cidrs" {
  type        = map(string)
  description = "CIDR blocks for the app subnets (EKS nodes and pods), keyed by availability zone."
  default = {
    "eu-west-2a" = "10.0.32.0/20"
    "eu-west-2b" = "10.0.48.0/20"
  }
}

variable "data_subnet_cidrs" {
  type        = map(string)
  description = "CIDR blocks for the data subnets (RDS), keyed by availability zone."
  default = {
    "eu-west-2a" = "10.0.21.0/24"
    "eu-west-2b" = "10.0.22.0/24"
  }
}

# EKS

variable "kubernetes_version" {
  type        = string
  description = "Kubernetes version for the EKS control plane (e.g., 1.37)."
  default     = "1.37"
}

variable "admin_role_arn" {
  type        = string
  description = "ARN of the IAM Identity Center admin role that gets cluster admin."
  default     = "arn:aws:iam::188735374190:role/aws-reserved/sso.amazonaws.com/eu-west-2/AWSReservedSSO_AdministratorAccess_ee638b5066f1f753"
}

variable "github_actions_role_arn" {
  type        = string
  description = "ARN of the GitHub Actions OIDC role that runs Terraform and bootstraps Argo CD."
  default     = "arn:aws:iam::188735374190:role/splitit-github-actions"
}

variable "node_instance_type" {
  type        = string
  description = "EC2 instance type for the managed node group (ARM, so images must be built for arm64)."
  default     = "t4g.medium"
}

variable "node_count" {
  type        = number
  description = "Fixed number of nodes in the managed node group."
  default     = 2
}

# Database

variable "db_engine_version" {
  type        = string
  description = "PostgreSQL major version (matches the local compose image)."
  default     = "16"
}

variable "db_instance_class" {
  type        = string
  description = "RDS instance class."
  default     = "db.t4g.micro"
}

variable "db_allocated_storage" {
  type        = number
  description = "Allocated storage in GB."
  default     = 20
}

variable "db_name" {
  type        = string
  description = "Name of the database created on the instance."
  default     = "splitit"
}

variable "db_username" {
  type        = string
  description = "Master username for the database."
  default     = "splitit"
}

# DNS

variable "hosted_zone_name" {
  type        = string
  description = "Existing Route 53 hosted zone for the app and add-ons (e.g., eks.example.com)."
  default     = "eks.hamzahsahal.com"
}

# Argo CD

variable "argocd_chart_version" {
  type        = string
  description = "Version of the argo-cd Helm chart."
  default     = "10.18.2"
}

variable "argocd_apps_chart_version" {
  type        = string
  description = "Version of the argocd-apps Helm chart that creates the root Application."
  default     = "2.0.8"
}
