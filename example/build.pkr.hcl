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

source "nebius" "foo-example" {
  mock = local.foo
}

source "nebius" "bar-example" {
  mock = local.bar
}

build {
  sources = [
    "source.nebius.foo-example",
  ]

  source "source.nebius.bar-example" {
    name = "bar"
  }

  provisioner "scaffolding-my-provisioner" {
    only = ["scaffolding-my-builder.foo-example"]
    mock = "foo: ${local.foo}"
  }

  provisioner "scaffolding-my-provisioner" {
    only = ["scaffolding-my-builder.bar"]
    mock = "bar: ${local.bar}"
  }

  post-processor "scaffolding-my-post-processor" {
    mock = "post-processor mock-config"
  }
}
