data "aws_availability_zones" "available" {}

resource "aws_vpc" "rke-vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = "${
    map(
     "Name", "${var.cluster_name}-rke-vpc",
     "kubernetes.io/cluster/${var.cluster_id}", "owned",
    )
  }"
}

resource "aws_subnet" "rke-subnet" {
  count = 1

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = "${aws_vpc.rke-vpc.id}"

  tags = "${
    map(
     "Name", "${var.cluster_name}-rke",
     "kubernetes.io/cluster/${var.cluster_id}", "owned",
    )
  }"
}

resource "aws_internet_gateway" "rke-ig" {
  vpc_id = "${aws_vpc.rke-vpc.id}"

  tags {
    Name = "${var.cluster_name}-rke-igw"
  }
}

resource "aws_route_table" "rke-rt" {
  vpc_id = "${aws_vpc.rke-vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.rke-ig.id}"
  }
}

resource "aws_route_table_association" "rke-rta" {
  count = 1

  subnet_id      = "${aws_subnet.rke-subnet.*.id[count.index]}"
  route_table_id = "${aws_route_table.rke-rt.id}"
}
