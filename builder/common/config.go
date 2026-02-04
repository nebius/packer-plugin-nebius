//go:generate packer-sdc mapstructure-to-hcl2 -type ServiceAccountConfig,DiskConfig,BaseImageConfig
package common

import (
	"fmt"
	"strings"
)

type ServiceAccountConfig struct {
	PrivateKeyFileEnv string `mapstructure:"private_key_file_env"`
	PublicKeyIDEnv    string `mapstructure:"public_key_id_env"`
	AccountIDEnv      string `mapstructure:"account_id_env"`
}

func (c *ServiceAccountConfig) Validate() error {
	if c.PrivateKeyFileEnv == "" {
		return fmt.Errorf("service_account.private_key_file_env is required")
	}
	if c.PublicKeyIDEnv == "" {
		return fmt.Errorf("service_account.public_key_file_env is required")
	}
	if c.AccountIDEnv == "" {
		return fmt.Errorf("service_account.account_id_env is required")
	}
	return nil
}

type DiskConfig struct {
	Type          string `mapstructure:"type"`
	SizeGibibytes int64  `mapstructure:"size_gibibytes"`
}

func (c *DiskConfig) Validate() error {
	diskType := strings.ToLower(strings.TrimSpace(c.Type))

	switch diskType {
	case "", "network_ssd", "network_ssd_non_replicated", "network_ssd_io_m3":
	default:
		return fmt.Errorf("invalid disk.type: %s", c.Type)
	}

	if c.SizeGibibytes < 10 {
		return fmt.Errorf("disk.size_gibibytes must be at least 10 GiB")
	}

	return nil
}

type BaseImageConfig struct {
	ID       string `mapstructure:"id"`
	ParentID string `mapstructure:"parent_id"`
	Family   string `mapstructure:"family"`
}

func (c *BaseImageConfig) Validate() error {
	if c.ID != "" {
		return nil
	}

	if c.ParentID == "" || c.Family == "" {
		return fmt.Errorf("either base_image.id must be set, or both base_image.parent_id and base_image.family must be set")
	}

	return nil
}
