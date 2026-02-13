# Copyright IBM Corp. 2020, 2025
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    nebius = {
      version = ">= 0.0.1"
      source  = "github.com/nebius/nebius"
    }
  }
}

source "nebius-image" "ubuntu2404-driverless" {
  communicator = "ssh"
  ssh_username = "ubuntu"
  service_account {
    private_key_file = "/path/to/private.pem"
    public_key_id    = "publickey-e0tk41vmw8sqsk6ja8"
    account_id       = "serviceaccount-e0tx3ejmbyn55rkfys"
  }
  disk {
    size_gibibytes = 10
  }
  base_image {
    family = "ubuntu24.04-driverless"
  }
  network {
    associate_public_ip_address = true
  }
  instance {
    platform = "cpu-d3"
    preset   = "4vcpu-16gb"
  }
  image {
    name                        = "ubuntu24.04-driverless-0.0.9"
    version                     = "0.0.9"
    image_family                = "ubuntu24.04-driverless"
    cpu_architecture            = "amd64"
    image_family_human_readable = "Ubuntu 24.04 Driverless"
  }
  parent_id = "project-e0tr8t9cc5s460k4r8n71"
}

build {
  sources = ["source.nebius-image.ubuntu2404-driverless"]
  provisioner "ansible" {
    playbook_file = "provision.yml"
  }
}
