locals {
  node_dns_list = "${aws_instance.rke-node.*.public_dns}"
}

data rke_node_parameter "nodes" {
  count   = "${var.node_count}"
  address = "${local.node_dns_list[count.index]}"
  user    = "ubuntu"
  role    = ["controlplane", "worker", "etcd"]
  ssh_key = "${tls_private_key.node-key.private_key_pem}"
}

resource rke_cluster "cluster" {
  nodes_conf = ["${data.rke_node_parameter.nodes.*.json}"]

  cloud_provider {
    name = "{var.cloud_provider}"
  }
}

###############################################################################
# If you need kubeconfig.yml for using kubectl, please uncomment follows.
###############################################################################
resource "local_file" "kube_cluster_yaml" {
  filename = "${path.root}/kube_config_cluster.yml"
  content  = "${rke_cluster.cluster.kube_config_yaml}"
}

resource "local_file" "rke_yaml" {
  filename = "${path.root}/rancher-cluster.yml"
  content = "${rke_cluster.cluster.rke_cluster_yaml}"
}
