output "vpc_cni_role_arn" {
  value       = module.vpc_cni_irsa.arn
  description = "ARN of the IAM role for the VPC CNI (aws-node)."
}

output "load_balancer_controller_role_arn" {
  value       = module.load_balancer_controller_irsa.arn
  description = "ARN of the IAM role for the AWS Load Balancer Controller."
}

output "external_dns_role_arn" {
  value       = module.external_dns_irsa.arn
  description = "ARN of the IAM role for ExternalDNS."
}

output "cert_manager_role_arn" {
  value       = module.cert_manager_irsa.arn
  description = "ARN of the IAM role for cert-manager."
}

output "external_secrets_role_arn" {
  value       = module.external_secrets_irsa.arn
  description = "ARN of the IAM role for External Secrets."
}
