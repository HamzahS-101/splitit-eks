output "vpc_id" {
  value       = aws_vpc.vpc.id
  description = "ID of the VPC."
}

output "public_subnet_ids" {
  value       = [for subnet in aws_subnet.public_subnet : subnet.id]
  description = "IDs of the public subnets."
}

output "app_subnet_ids" {
  value       = [for subnet in aws_subnet.app_subnet : subnet.id]
  description = "IDs of the app subnets."
}

output "data_subnet_ids" {
  value       = [for subnet in aws_subnet.data_subnet : subnet.id]
  description = "IDs of the data subnets."
}
