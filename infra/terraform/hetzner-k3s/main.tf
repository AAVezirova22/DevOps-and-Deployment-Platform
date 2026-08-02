terraform {
  required_version = ">= 1.6.0"

  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.49"
    }
  }
}

provider "hcloud" {
  token = var.hcloud_token
}

resource "hcloud_ssh_key" "operator" {
  name       = "${var.project_name}-operator"
  public_key = var.ssh_public_key
}

resource "hcloud_firewall" "k3s" {
  name = "${var.project_name}-k3s"

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "22"
    source_ips = var.admin_cidrs
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "80"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "443"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "6443"
    source_ips = var.admin_cidrs
  }
}

resource "hcloud_server" "k3s" {
  name        = "${var.project_name}-k3s-1"
  image       = "ubuntu-24.04"
  server_type = var.server_type
  location    = var.server_location
  ssh_keys    = [hcloud_ssh_key.operator.id]
  firewall_ids = [
    hcloud_firewall.k3s.id
  ]

  user_data = templatefile("${path.module}/cloud-init.yaml", {
    k3s_channel = var.k3s_channel
  })

  labels = {
    project = var.project_name
    role    = "k3s"
  }
}
