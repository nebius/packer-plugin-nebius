//go:generate packer-sdc mapstructure-to-hcl2 -type Config
package instance

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"

	nebiuscommon "github.com/hashicorp/packer-plugin-nebius/builder/common"
)

type Config struct {
	common.PackerConfig  `mapstructure:",squash"`
	ParentID             string                            `mapstructure:"parent_id"`
	ServiceAccountConfig nebiuscommon.ServiceAccountConfig `mapstructure:"service_account"`
	DiskConfig           nebiuscommon.DiskConfig           `mapstructure:"disk"`
	BaseImageConfig      nebiuscommon.BaseImageConfig      `mapstructure:"base_image"`
	NetworkConfig        nebiuscommon.NetworkConfig        `mapstructure:"network"`
	InstanceConfig       nebiuscommon.InstanceConfig       `mapstructure:"instance"`
}

func (c *Config) validate() error {
	if c.ParentID == "" {
		return fmt.Errorf("parent_id is required")
	}

	if err := c.ServiceAccountConfig.Validate(); err != nil {
		return err
	}
	if err := c.DiskConfig.Validate(); err != nil {
		return err
	}
	if err := c.BaseImageConfig.Validate(); err != nil {
		return err
	}
	if err := c.NetworkConfig.Validate(); err != nil {
		return err
	}

	return nil
}
