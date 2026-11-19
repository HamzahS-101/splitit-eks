output "cluster_name" {
  value       = module.eks.cluster_name
  description = "Name of the EKS cluster."
}

output "cluster_endpoint" {
  value       = module.eks.cluster_endpoint
  description = "URL of the Kubernetes API server."
}

output "cluster_certificate_authority_data" {
  value       = module.eks.cluster_certificate_authority_data
  description = "Base64 CA certificate used to verify the API server."
}

output "oidc_provider_arn" {
  value       = module.eks.oidc_provider_arn
  description = "ARN of the cluster's IAM OIDC provider, used by IRSA roles."
}

output "node_security_group_id" {
  value       = module.eks.node_security_group_id
  description = "ID of the node security group, used by the RDS security group."
}
