### Installation

Add the plugin to your Packer configuration, then run `packer init`:

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

You can also install a local build:

```sh
packer plugins install --path packer-plugin-nebius github.com/nebius/nebius
```

### [How to use with Nebius Cloud](https://docs.nebius.com/compute/storage/packer#creating-boot-disk-images-with-packer)

### Components

#### Builders

- `nebius-image` - Builds an image by creating a VM, provisioning it over SSH, and publishing either the boot disk or an attached secondary disk.
