output "db_instance_id" {
  value       = aws_db_instance.db.identifier
  description = "Identifier of the RDS instance."
}

output "db_password_secret_arn" {
  value       = aws_db_instance.db.master_user_secret[0].secret_arn
  description = "ARN of the RDS-managed Secrets Manager secret holding the master username and password."
}

output "db_host_parameter_arn" {
  value       = aws_ssm_parameter.db_host.arn
  description = "ARN of the SSM parameter holding the RDS hostname."
}
