[![acceptance tests](https://github.com/nebius/packer-plugin-nebius/actions/workflows/testacc.yml/badge.svg)](https://github.com/nebius/packer-plugin-nebius/actions/workflows/testacc.yml)

# Packer Plugin Nebius

Packer Plugin Nebius provides a Nebius Compute builder for creating custom images from base images. The plugin is designed to integrate cleanly into standard Packer workflows via `packer init` and the required plugin block. Configuration focuses on explicit control of base images, instance shape, and image metadata. The builder is optimized for repeatable image pipelines in Nebius projects. Example usage is included to help you get started quickly. © Nebius BV, 2026.

## Installation

Add the plugin to your Packer configuration and run `packer init`:

```hcl
packer {
  required_plugins {
    nebius = {
      source  = "github.com/nebius/nebius"
      version = ">= 0.0.1"
    }
  }
}
```

Or install a local build:

```sh
packer plugins install --path packer-plugin-nebius github.com/nebius/nebius
```

## Builder: `nebius-image`

Key settings:
- `parent_id` - Project or folder to place resources in.
- `service_account` - `public_key_id`, `account_id`, and one of `private_key` or `private_key_file`.
- `base_image` - `id` or `family`.
- `disk` - `size_gibibytes` (minimum 10), optional `type`.
- `network` - `associate_public_ip_address` (optional, auto allocation) or `public_allocation_id` (optional, preallocated public ID).
- `instance` - `platform` and `preset`.
- `image` - `name` (required), optional family metadata.
- `ssh_username` - required; only `ssh` communicator is supported.

Example is available in `example/build.pkr.hcl`.

## Build from source

```sh
go build -ldflags="-X github.com/hashicorp/packer-plugin-nebius/version.VersionPrerelease=dev" -o packer-plugin-nebius
```

## Testing

```sh
PKR_VAR_nb_parent_id=project_id PKR_VAR_nb_token=token make testacc
```

## Disclaimer

packer-plugin-nebius is not created nor endorsed by HashiCorp or IBM Corporation.
Nebius B.V. is not affiliated with HashiCorp or IBM Corporation.
