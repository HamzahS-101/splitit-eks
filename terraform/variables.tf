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

variable "admin_cidrs" {
  type        = list(string)
  description = "CIDR blocks allowed to reach the public API endpoint. Set with TF_VAR_admin_cidrs so your IP stays out of Git."
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
