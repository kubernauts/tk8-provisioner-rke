# Variables Configuration
variable "cluster_name" {
  default     = "kubernauts"
  type        = "string"
  description = "The name of your EKS Cluster"
}

variable "aws_region" {
  default = "us-east-1"

  # availabe regions are:
  # us-east-1 (Virginia)
  # us-west-2 (Oregon)
  # eu-west-1 (Irland)
  type = "string"

  description = "The AWS Region to deploy EKS"
}

variable "rke_node_instance_type" {
  default     = "t2.medium"
  type        = "string"
  description = "Node EC2 instance type"
}

variable "node_count" {
  default     = 3
  type        = "string"
  description = "Autoscaling Desired node capacity"
}

variable "authorization" {
  default     = "rbac"
  type        = "string"
  description = "authorization mode in rke cluster"
}

variable "AWS_ACCESS_KEY_ID" {
  description = "AWS Access Key"
}

variable "AWS_SECRET_ACCESS_KEY" {
  description = "AWS Secret Key"
}

variable "AWS_DEFAULT_REGION" {
  description = "AWS Region"
}

variable "cloud_provider" {
  default     = "aws"
  description = "cloud provider for rke"
}
