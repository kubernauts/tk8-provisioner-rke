# RKE Terraform module

module "rke" {
  source                 = "./modules/rke"
  cluster_name           = "${var.cluster_name}"
  aws_region             = "${var.aws_region}"
  rke_node_instance_type = "${var.rke_node_instance_type}"
  node_count             = "${var.node_count}"
  cloud_provider         = "${var.cloud_provider}"
  AWS_ACCESS_KEY_ID      = "${var.AWS_ACCESS_KEY_ID}"
  AWS_SECRET_ACCESS_KEY  = "${var.AWS_SECRET_ACCESS_KEY}"
  AWS_DEFAULT_REGION     = "${var.AWS_DEFAULT_REGION}"
}
