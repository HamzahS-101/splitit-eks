output "db_instance_id" {
  value       = aws_db_instance.db.identifier
  description = "Identifier of the RDS instance."
}

output "db_password_secret_arn" {
  value       = aws_secretsmanager_secret.db_password.arn
  description = "ARN of the Secrets Manager secret holding the master password."
}

output "db_host_parameter_arn" {
  value       = aws_ssm_parameter.db_host.arn
  description = "ARN of the SSM parameter holding the RDS hostname."
}
