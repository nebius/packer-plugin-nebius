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

### Components

#### Builders

- `nebius-image` - Builds an image by creating a VM, provisioning it over SSH, and publishing the result.
