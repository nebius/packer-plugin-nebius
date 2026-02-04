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

source "nebius-instance" "auth-check" {
  service_account {
    private_key_file_env = "NB_AUTHKEY_PRIVATE_PATH"
    public_key_id_env    = "NB_AUTHKEY_PUBLIC_ID"
    account_id_env       = "NB_SA_ID"
  }
  disk {
    size_gibibytes = 10
  }
  base_image {
    id = "computeimage-e0tnmenkcw3exfx4mm"
  }
  parent_id = "project-e0tr8t9cc5s460k4r8n71"
}

build {
  sources = ["source.nebius-instance.auth-check"]
}