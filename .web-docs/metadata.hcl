# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Details on using this Integration template can be found at https://github.com/hashicorp/integration-template
# This metadata.hcl file and the adjacent `components` docs directory should
# be kept in a `.web-docs` directory at the root of your plugin repository.
integration {
  name = "Nebius"
  description = "Packer builder for Nebius Compute that creates a VM from a base image, provisions it over SSH, and publishes a new image."
  identifier = "packer/nebius/nebius"
  docs {
    process_docs = true
    readme_location = "./README.md"
    external_url = "https://github.com/nebius/packer-plugin-nebius"
  }
  license {
    type = "MPL-2.0"
    url = "https://github.com/nebius/packer-plugin-nebius/blob/main/LICENSE"
  }
  component {
    type = "builder"
    name = "Nebius Image"
    slug = "nebius-image"
  }
}
