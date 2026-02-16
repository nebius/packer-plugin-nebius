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
    # If you'd prefer not to publish docs on HashiCorp websites, you can
    # set `process_docs` to `false`. If `process_docs` is `false`, you MUST
    # provide a `external_url` so we can link back to your plugin repo.
    process_docs = true
    # Note that the README location is relative to this file. We recommend
    # keeping the default value, as the adjacent `compile-to-webdocs` script
    # will automatically copy the README from the `docs` directory of this
    # repository to the correct location.
    readme_location = "./README.md"
    # `external_url` allows us to link back to your plugin repo.
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
