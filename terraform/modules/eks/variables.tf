variable "project_name" {
  type        = string
  description = "Project name, used as a prefix for resource names."
}

variable "kubernetes_version" {
  type        = string
  description = "Kubernetes version for the EKS control plane (e.g., 1.37)."
}

variable "vpc_id" {
  type        = string
  description = "ID of the VPC the cluster runs in."
}

variable "app_subnet_ids" {
  type        = list(string)
  description = "IDs of the private subnets for the control plane ENIs and the nodes."
}

variable "admin_role_arn" {
  type        = string
  description = "ARN of the IAM role (SSO admin) given cluster admin through an access entry."
}

variable "github_actions_role_arn" {
  type        = string
  description = "ARN of the GitHub Actions OIDC role given cluster admin to bootstrap Argo CD."
}

variable "node_instance_type" {
  type        = string
  description = "EC2 instance type for the managed node group."
}

variable "node_count" {
  type        = number
  description = "Fixed number of nodes in the managed node group."
}

variable "vpc_cni_role_arn" {
  type        = string
  description = "ARN of the IRSA role the VPC CNI (aws-node) uses instead of the node role."
}
