variable "project_name" {
  type        = string
  description = "Project name, used as a prefix for resource names."
}

variable "vpc_id" {
  type        = string
  description = "ID of the VPC the RDS security group is created in."
}

variable "data_subnet_ids" {
  type        = list(string)
  description = "Data subnet IDs for the DB subnet group."
}

variable "node_sg_id" {
  type        = string
  description = "ID of the EKS node security group allowed to reach PostgreSQL."
}

variable "db_engine_version" {
  type        = string
  description = "PostgreSQL major version."
}

variable "db_instance_class" {
  type        = string
  description = "RDS instance class."
}

variable "db_allocated_storage" {
  type        = number
  description = "Allocated storage in GB."
}

variable "db_name" {
  type        = string
  description = "Name of the database created on the instance."
}

variable "db_username" {
  type        = string
  description = "Master username for the database."
}
