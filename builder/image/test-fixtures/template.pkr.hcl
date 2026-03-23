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

variable "NB_PARENT_ID" {
  type = string
}

variable "NB_PUB_KEY" {
  type = string
}

variable "NB_SA" {
  type = string
}

variable "NB_PRIVATE_KEY" {
  type = string
}

locals {
  nb_image_name = "acc-${uuidv4()}"
}

source "nebius-image" "acceptance" {
  communicator = "ssh"
  ssh_username = "ubuntu"
  parent_id = var.NB_PARENT_ID

  service_account {
    account_id = var.NB_SA
    public_key_id = var.NB_PUB_KEY
    private_key = var.NB_PRIVATE_KEY
  }

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
