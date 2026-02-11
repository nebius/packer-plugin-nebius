# Copyright IBM Corp. 2020, 2025
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    nebius = {
      version = ">= 0.0.1"
      source  = "gitlab.nebius.dev/project-compute/nebius"
    }
  }
}

source "nebius-instance" "image-create" {
  communicator = "ssh"
  service_account {
    private_key_file_env = "NB_AUTHKEY_PRIVATE_PATH"
    public_key_id_env    = "NB_AUTHKEY_PUBLIC_ID"
    account_id_env       = "NB_SA_ID"
    private_key_file     = "/Users/ruslan/projects/nebius/packer-plugin-nebius/example/authkey/private.pem"
    public_key_id        = "publickey-e0tk41vmw8sqsk6ja8"
    account_id           = "serviceaccount-e0tx3ejmbyn55rkfys"
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
    name = "ubuntu24.04-driverless-0.0.4"
    version = "0.0.4"
    image_family = "ubuntu24.04-driverless-wolfwalker"
    cpu_architecture = "amd64"
    image_family_human_readable = "Ubuntu 24.04 Driverless"
  }
  parent_id = "project-e0tr8t9cc5s460k4r8n71"
}

build {
  sources = ["source.nebius-instance.image-create"]
  provisioner "ansible" {
    playbook_file = "provision.yml"
  }
}