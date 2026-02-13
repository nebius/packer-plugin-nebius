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

variable "nb_private_key_file" {
  type = string
}

variable "nb_public_key_id" {
  type = string
}

variable "nb_sa_id" {
  type = string
}

variable "nb_parent_id" {
  type = string
}

variable "nb_base_image_family" {
  type = string
}

variable "nb_platform" {
  type = string
}

variable "nb_preset" {
  type = string
}

variable "nb_image_name" {
  type = string
}

variable "nb_image_version" {
  type = string
}

variable "nb_image_family" {
  type = string
}

variable "nb_image_family_human_readable" {
  type = string
}

variable "nb_cpu_architecture" {
  type = string
}

variable "nb_api_endpoint" {
  type    = string
  default = ""
}

source "nebius-image" "acceptance" {
  api_endpoint = var.nb_api_endpoint
  communicator = "ssh"
  ssh_username = "ubuntu"

  service_account {
    private_key_file = var.nb_private_key_file
    public_key_id    = var.nb_public_key_id
    account_id       = var.nb_sa_id
  }

  parent_id = var.nb_parent_id

  base_image {
    family = var.nb_base_image_family
  }

  disk {
    size_gibibytes = 10
  }

  network {
    associate_public_ip_address = true
  }

  instance {
    platform = var.nb_platform
    preset   = var.nb_preset
  }

  image {
    name                        = var.nb_image_name
    version                     = var.nb_image_version
    image_family                = var.nb_image_family
    image_family_human_readable = var.nb_image_family_human_readable
    cpu_architecture            = var.nb_cpu_architecture
  }
}

build {
  sources = ["source.nebius-image.acceptance"]
}
