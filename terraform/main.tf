module "network" {
  source = "./modules/network"

  project_name        = var.project_name
  vpc_cidr            = var.vpc_cidr
  public_subnet_cidrs = var.public_subnet_cidrs
  app_subnet_cidrs    = var.app_subnet_cidrs
  data_subnet_cidrs   = var.data_subnet_cidrs
}

module "eks" {
  source = "./modules/eks"

  project_name       = var.project_name
  kubernetes_version = var.kubernetes_version
  vpc_id             = module.network.vpc_id
  app_subnet_ids     = module.network.app_subnet_ids
  admin_cidrs        = var.admin_cidrs
  node_instance_type = var.node_instance_type
  node_count         = var.node_count
  vpc_cni_role_arn   = module.irsa.vpc_cni_role_arn
}

module "database" {
  source = "./modules/database"

  project_name         = var.project_name
  vpc_id               = module.network.vpc_id
  data_subnet_ids      = module.network.data_subnet_ids
  node_sg_id           = module.eks.node_security_group_id
  db_engine_version    = var.db_engine_version
  db_instance_class    = var.db_instance_class
  db_allocated_storage = var.db_allocated_storage
  db_name              = var.db_name
  db_username          = var.db_username
}

data "aws_route53_zone" "zone" {
  name = var.hosted_zone_name
}

module "irsa" {
  source = "./modules/irsa"

  project_name           = var.project_name
  oidc_provider_arn      = module.eks.oidc_provider_arn
  hosted_zone_arn        = data.aws_route53_zone.zone.arn
  db_password_secret_arn = module.database.db_password_secret_arn
  db_host_parameter_arn  = module.database.db_host_parameter_arn
}
