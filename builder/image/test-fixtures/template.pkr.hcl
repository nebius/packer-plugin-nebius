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

variable "nb_parent_id" {
  type = string
}

variable "nb_token" {
  type = string
}

locals {
  nb_image_name = "acc-${uuidv4()}"
}

source "nebius-image" "acceptance" {
  api_endpoint = "api.testing.nebius.cloud:443"
  communicator = "ssh"
  ssh_username = "ubuntu"
  parent_id = var.nb_parent_id
  token = var.nb_token

  base_image {
    family = "ubuntu24.04-driverless"
  }

  disk {
    size_gibibytes = 10
  }

  network {
    associate_public_ip_address = true
  }

  instance {
    platform = "cpu-d3"
    preset   = "4vcpu-16gb"
  }

  image {
    name = local.nb_image_name
  }
}

build {
  sources = ["source.nebius-image.acceptance"]
  provisioner "ansible" {
    playbook_file = "./test-fixtures/provision.yml"
  }
}
