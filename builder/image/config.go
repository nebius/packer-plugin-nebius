//go:generate packer-sdc mapstructure-to-hcl2 -type Config
package image

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/uuid"

	nebiuscommon "github.com/hashicorp/packer-plugin-nebius/builder/common"
)

type Config struct {
	common.PackerConfig  `mapstructure:",squash"`
	Comm                 communicator.Config               `mapstructure:",squash"`
	UseSecondaryDisk     bool                              `mapstructure:"use_secondary_disk"`
	APIEndpoint          string                            `mapstructure:"api_endpoint"`
	ParentID             string                            `mapstructure:"parent_id"`
	Token                string                            `mapstructure:"token"`
	ServiceAccountConfig nebiuscommon.ServiceAccountConfig `mapstructure:"service_account"`
	DiskConfig           nebiuscommon.DiskConfig           `mapstructure:"disk"`
	SecondaryDiskConfig  nebiuscommon.SecondaryDiskConfig  `mapstructure:"secondary_disk"`
	BaseImageConfig      nebiuscommon.BaseImageConfig      `mapstructure:"base_image"`
	NetworkConfig        nebiuscommon.NetworkConfig        `mapstructure:"network"`
	InstanceConfig       nebiuscommon.InstanceConfig       `mapstructure:"instance"`
	ImageConfig          nebiuscommon.ImageConfig          `mapstructure:"image"`
}

func (c *Config) validate() []error {
	errors := []error{}

	if c.ParentID == "" {
		errors = append(errors, fmt.Errorf("parent_id must be set"))
	}

	if c.Token == "" {
		if err := c.ServiceAccountConfig.Validate(); err != nil {
			errors = append(errors, err)
		}
	}
	if err := c.DiskConfig.Validate(); err != nil {
		errors = append(errors, err)
	}
	if c.UseSecondaryDisk {
		if err := c.SecondaryDiskConfig.Validate(); err != nil {
			errors = append(errors, fmt.Errorf("secondary_disk: %w", err))
		}
	} else if c.SecondaryDiskConfig != (nebiuscommon.SecondaryDiskConfig{}) {
		errors = append(errors, fmt.Errorf("secondary_disk requires use_secondary_disk = true"))
	}
	if err := c.BaseImageConfig.Validate(); err != nil {
		errors = append(errors, err)
	}
	if err := c.NetworkConfig.Validate(); err != nil {
		errors = append(errors, err)
	}
	if err := c.InstanceConfig.Validate(); err != nil {
		errors = append(errors, err)
	}
	if err := c.ImageConfig.Validate(); err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *Config) prepareSSH(ctx *interpolate.Context) []error {
	if c.Comm.Type != "" && c.Comm.Type != "ssh" {
		return []error{fmt.Errorf("unsupported communicator type: %s", c.Comm.Type)}
	}

	if c.Comm.SSH.SSHUsername == "" {
		return []error{fmt.Errorf("ssh_username must be set for SSH communicator")}
	}

	if c.Comm.SSH.SSHTemporaryKeyPairType != "" && c.Comm.SSHTemporaryKeyPairType != "rsa" && c.Comm.SSH.SSHTemporaryKeyPairType != "ed25519" {
		return []error{fmt.Errorf("temporary_key_pair_type requires either rsa or ed25519 as its value")}
	}

	if c.Comm.SSHKeyPairName != "" {
		if c.Comm.SSHPrivateKeyFile == "" && !c.Comm.SSH.SSHAgentAuth {
			return []error{fmt.Errorf("ssh_private_key_file must be provided or ssh_agent_auth enabled when ssh_keypair_name is specified")}
		}
	}

	// If we are not given an explicit ssh_keypair_name or
	// ssh_private_key_file, then create a temporary one, but only if the
	// temporary_key_pair_name has not been provided, and we are not using
	// ssh_password.
	if c.Comm.SSHKeyPairName == "" && c.Comm.SSH.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSH.SSHPrivateKeyFile == "" && c.Comm.SSH.SSHPassword == "" {
		c.Comm.SSH.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	if c.Comm.SSH.SSHTemporaryKeyPairType == "" {
		c.Comm.SSH.SSHTemporaryKeyPairType = "ed25519"
	}

	if c.Comm.SSHPrivateKeyFile != "" {
		// Using existing SSH private key
		privateKeyBytes, err := c.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			return []error{fmt.Errorf("failed to read ssh private key file: %w", err)}
		}

		c.Comm.SSHPrivateKey = privateKeyBytes
		return nil
	}

	if c.Comm.SSHAgentAuth && c.Comm.SSHKeyPairName == "" {
		// Using SSH Agent with key pair in Source AMI
		return nil
	}

	if c.Comm.SSHAgentAuth && c.Comm.SSHKeyPairName != "" {
		// "Using SSH Agent for existing key pair %s", s.Comm.SSHKeyPairName))
		return nil
	}

	if c.Comm.SSHTemporaryKeyPairName == "" {
		// "Not using temporary keypair"
		c.Comm.SSHKeyPairName = ""
		return nil
	}

	return c.Comm.Prepare(ctx)
}
