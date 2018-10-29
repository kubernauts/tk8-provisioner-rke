locals {
  cluster_id_tag = "${map("kubernetes.io/cluster/${var.cluster_id}", "owned")}"
}

resource "aws_security_group" "rke_security_group" {
  name        = "rke-sg-1"
  description = "Allow inbound/outbound traffic for rke"
  vpc_id      = "${aws_vpc.rke-vpc.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = "${local.cluster_id_tag}"
}

resource "aws_security_group_rule" "ingress_test" {
  type              = "ingress"
  protocol          = "tcp"
  from_port         = "2379"
  to_port           = "2380"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.rke_security_group.id}"
}

resource "aws_security_group_rule" "ingress_test1" {
  type              = "ingress"
  protocol          = "tcp"
  from_port         = "10250"
  to_port           = "10252"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.rke_security_group.id}"
}

resource "aws_security_group_rule" "ingress_world" {
  count = "${length(var.aws_rke_sg_ing_defaults)}"

  type        = "ingress"
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]
  from_port   = "${element(var.aws_rke_sg_ing_defaults, count.index)}"
  to_port     = "${element(var.aws_rke_sg_ing_defaults, count.index)}"

  security_group_id = "${aws_security_group.rke_security_group.id}"
}

resource "aws_security_group_rule" "ingress_self" {
  count = "${length(split(",", var.aws_rke_sg_ing_self1["protocol"]))}"

  type      = "ingress"
  protocol  = "${element(split(",", var.aws_rke_sg_ing_self1["protocol"]), count.index)}"
  self      = true
  from_port = "${element(split(",", var.aws_rke_sg_ing_self1["from"]), count.index)}"
  to_port   = "${element(split(",", var.aws_rke_sg_ing_self1["to"]), count.index)}"

  security_group_id = "${aws_security_group.rke_security_group.id}"
}

resource "aws_instance" "rke-node" {
  count      = "${var.node_count}"
  depends_on = ["aws_internet_gateway.rke-ig", "aws_security_group.rke_security_group", "aws_security_group_rule.ingress_test", "aws_security_group_rule.ingress_test1", "aws_security_group_rule.ingress_world", "aws_security_group_rule.ingress_self"]

  ami                         = "${data.aws_ami.distro.id}"
  instance_type               = "${var.rke_node_instance_type}"
  key_name                    = "${aws_key_pair.rke-node-key.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.rke-aws.name}"
  subnet_id                   = "${aws_subnet.rke-subnet.id}"
  vpc_security_group_ids      = ["${aws_security_group.rke_security_group.id}"]
  associate_public_ip_address = true
  tags                        = "${local.cluster_id_tag}"

  provisioner "remote-exec" {
    connection {
      user        = "ubuntu"
      private_key = "${tls_private_key.node-key.private_key_pem}"
    }

    inline = [
      "curl https://releases.rancher.com/install-docker/17.03.sh | sh",
      "sudo usermod -a -G docker ubuntu",
    ]
  }
}

resource "aws_lb_target_group" "rancher-target-groups" {
  count       = "${length(var.aws_nlb_name)}"
  name        = "${element(var.aws_nlb_name, count.index)}"
  port        = "${element(var.aws_nlb_port, count.index)}"
  protocol    = "${var.aws_nlb_protocol}"
  target_type = "${var.aws_nlb_target_type}"
  vpc_id      = "${aws_vpc.rke-vpc.id}"

  health_check {
    interval            = "${var.aws_nlb_health_check_defaults["health_check_interval"]}"
    path                = "${var.aws_nlb_health_check_defaults["health_check_path"]}"
    port                = "${var.aws_nlb_health_check_defaults["health_check_port"]}"
    protocol            = "${var.aws_nlb_health_check_defaults["health_check_protocol"]}"
    healthy_threshold   = "${var.aws_nlb_health_check_defaults["health_check_healthy_threshold"]}"
    unhealthy_threshold = "${var.aws_nlb_health_check_defaults["health_check_unhealthy_threshold"]}"
    timeout             = "${var.aws_nlb_health_check_defaults["health_check_timeout"]}"
  }
}

resource "aws_lb_target_group_attachment" "attach_nodes_rancher_tg-80" {
  count            = "${aws_instance.rke-node.count}"
  target_group_arn = "${aws_lb_target_group.rancher-target-groups.0.arn}"
  target_id        = "${element(aws_instance.rke-node.*.id, count.index)}"
}

resource "aws_lb_target_group_attachment" "attach_nodes_rancher_tg-443" {
  count            = "${aws_instance.rke-node.count}"
  target_group_arn = "${aws_lb_target_group.rancher-target-groups.1.arn}"
  target_id        = "${element(aws_instance.rke-node.*.id, count.index)}"
}

resource "aws_lb" "test" {
  name               = "rancher-lb-tf"
  internal           = false
  load_balancer_type = "network"
  subnets            = ["${aws_subnet.rke-subnet.id}"]

  enable_deletion_protection = false

  tags {
    Environment = "production"
  }
}

resource "aws_lb_listener" "attach-tg-rke-nlb-443" {
  load_balancer_arn = "${aws_lb.test.arn}"
  port              = "443"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.rancher-target-groups.0.arn}"
  }
}

resource "aws_lb_listener" "attach-tg-rke-nlb-80" {
  load_balancer_arn = "${aws_lb.test.arn}"
  port              = "80"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.rancher-target-groups.1.arn}"
  }
}
