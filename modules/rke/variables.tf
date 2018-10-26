variable "aws_rke_sg_ing_defaults" {
  description = "Inbound rules for rke security group"
  type        = "list"
  default     = ["80", "6443", "22", "2376", "443"]
}

variable "aws_rke_sg_ing_self1" {
  description = "Inbound rules for rke security group"
  type        = "map"

  default = {
    "protocol" = "tcp,tcp,udp,udp,udp"
    "from"     = "30000,8472,30000,4789,10256"
    "to"       = "32767,8472,32767,4789,10256"
  }
}

variable "aws_nlb_name" {
  description = "NLB name"
  type        = "list"
  default     = ["rancher1-tcp-443", "rancher1-tcp-80"]
}

variable "aws_nlb_port" {
  description = "NLB ports"
  type        = "list"
  default     = ["443", "80"]
}

variable "aws_nlb_protocol" {
  description = "NLB protocol"
  default     = "TCP"
}

variable "aws_nlb_target_type" {
  description = "NLB target type"
  default     = "instance"
}

variable "aws_nlb_health_check_defaults" {
  description = "Default values for NLB health checks"
  type        = "map"

  default = {
    "health_check_interval"            = 10
    "health_check_healthy_threshold"   = 3
    "health_check_path"                = "/healthz"
    "health_check_port"                = "80"
    "health_check_protocol"            = "HTTP"
    "health_check_timeout"             = 6
    "health_check_healthy_threshold"   = 3
    "health_check_unhealthy_threshold" = 3
    "target_type"                      = "instance"
  }
}

variable "rke_node_instance_type" {
  type = "string"
}

variable "ssh_key_path" {
  type = "string"
}

variable "cluster_name" {
  type = "string"
}

variable "aws_region" {
  type = "string"
}

variable "node_count" {
  type = "string"
}

variable "AWS_SECRET_ACCESS_KEY" {
  type = "string"
}

variable "AWS_DEFAULT_REGION" {
  type = "string"
}

variable "AWS_ACCESS_KEY_ID" {
  type = "string"
}

variable "cluster_id" {
  default = "rke"
}

variable "cloud_provider" {
  type = "string"
}
