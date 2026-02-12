//go:generate packer-sdc mapstructure-to-hcl2 -type ServiceAccountConfig,DiskConfig,BaseImageConfig,NetworkConfig,InstanceConfig,ImageConfig
package common

import (
	"fmt"
	"slices"
	"strings"
)

type ServiceAccountConfig struct {
	PrivateKeyFileEnv string `mapstructure:"private_key_file_env"`
	PublicKeyIDEnv    string `mapstructure:"public_key_id_env"`
	AccountIDEnv      string `mapstructure:"account_id_env"`
	PrivateKeyFile    string `mapstructure:"private_key_file"`
	PublicKeyID       string `mapstructure:"public_key_id"`
	AccountID         string `mapstructure:"account_id"`
}

func (c *ServiceAccountConfig) Validate() error {
	if c.PrivateKeyFileEnv == "" && c.PrivateKeyFile == "" {
		return fmt.Errorf("either service_account.private_key_file_env or service_account.private_key_file must be set")
	}
	if c.PublicKeyIDEnv == "" && c.PublicKeyID == "" {
		return fmt.Errorf("either service_account.public_key_id_env or service_account.public_key_id must be set")
	}
	if c.AccountIDEnv == "" && c.AccountID == "" {
		return fmt.Errorf("either service_account.account_id_env or service_account.account_id must be set")
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
	if c.ID == "" && c.Family == "" {
		return fmt.Errorf("base_image.id or base_image.family must be set")
	}

	return nil
}

type NetworkConfig struct {
	SubnetID                 string `mapstructure:"subnet_id"`
	AssociatePublicIpAddress bool   `mapstructure:"associate_public_ip_address"`
}

func (c *NetworkConfig) Validate() error {
	return nil
}

type InstanceConfig struct {
	Platform string `mapstructure:"platform"`
	Preset   string `mapstructure:"preset"`
}

func (c *InstanceConfig) Validate() error {
	if c.Platform == "" {
		return fmt.Errorf("instance.platform is required")
	}
	if c.Preset == "" {
		return fmt.Errorf("instance.preset is required")
	}
	return nil
}

type ImageConfig struct {
	Name                     string            `mapstructure:"name"`
	Labels                   map[string]string `mapstructure:"labels"`
	ImageFamily              string            `mapstructure:"image_family"`
	ImageFamilyHumanReadable string            `mapstructure:"image_family_human_readable"`
	Version                  string            `mapstructure:"version"`
	CPUArchitecture          string            `mapstructure:"cpu_architecture"`
	ParentID                 string            `mapstructure:"parent_id"`
	UnsupportedPlatforms     map[string]string `mapstructure:"unsupported_platforms"`
	RecommendedPlatforms     []string          `mapstructure:"recommended_platforms"`
}

func (c *ImageConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("image.name is required")
	}
	if c.ImageFamily != "" && c.Version == "" {
		return fmt.Errorf("image.version is required when image.image_family is set")
	}
	if c.ImageFamily != "" && c.ImageFamilyHumanReadable == "" {
		return fmt.Errorf("image.image_family_human_readable is required when image.image_family is set")
	}

	normalizedCPUArch := strings.ToLower(strings.TrimSpace(c.CPUArchitecture))
	if !slices.Contains([]string{"", "arm64", "amd64"}, normalizedCPUArch) {
		return fmt.Errorf("invalid image.cpu_architecture: %s. use one of: arm64, amd64", c.CPUArchitecture)
	}

	return nil
}
