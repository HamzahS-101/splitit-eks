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
}
