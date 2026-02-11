package instance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/packer-plugin-nebius/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/nebius/gosdk"
	commonv1 "github.com/nebius/gosdk/proto/nebius/common/v1"
	computev1 "github.com/nebius/gosdk/proto/nebius/compute/v1"
)

const stateInstanceID = "instance_id"
const stateIPAddress = "ip_address"

type StepCreateInstance struct {
	sdk    *gosdk.SDK
	config *Config
}

func NewStepCreateInstance(sdk *gosdk.SDK, config *Config) *StepCreateInstance {
	return &StepCreateInstance{
		sdk:    sdk,
		config: config,
	}
}

func (s *StepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Message("Creating instance...")

	diskID := state.Get(stateDiskID).(string)
	subnetID := state.Get(stateSubnetID).(string)

	var publicIPAddress *computev1.PublicIPAddress
	if s.config.NetworkConfig.AssociatePublicIpAddress {
		publicIPAddress = &computev1.PublicIPAddress{}
	}

	cloudInitUserData := s.BuildUserData()
	if s.config.PackerDebug {
		ui.Message(fmt.Sprintf("Cloud-init user data:\n%s", cloudInitUserData))
	}

	req := &computev1.CreateInstanceRequest{
		Metadata: &commonv1.ResourceMetadata{
			Name: fmt.Sprintf("packer-%s-instance", uuid.New().String()),
			Labels: map[string]string{
				"packer_build": s.config.PackerBuildName,
			},
		},
		Spec: &computev1.InstanceSpec{
			Resources: &computev1.ResourcesSpec{
				Platform: s.config.InstanceConfig.Platform,
				Size: &computev1.ResourcesSpec_Preset{
					Preset: s.config.InstanceConfig.Preset,
				},
			},
			NetworkInterfaces: []*computev1.NetworkInterfaceSpec{
				{
					SubnetId:        subnetID,
					Name:            fmt.Sprintf("packer-%s-interface", uuid.New().String()),
					IpAddress:       &computev1.IPAddress{},
					PublicIpAddress: publicIPAddress,
				},
			},
			BootDisk: &computev1.AttachedDiskSpec{
				AttachMode: computev1.AttachedDiskSpec_READ_WRITE,
				Type: &computev1.AttachedDiskSpec_ExistingDisk{
					ExistingDisk: &computev1.ExistingDisk{
						Id: diskID,
					},
				},
			},
			CloudInitUserData: s.BuildUserData(),
		},
	}

	resp, err := s.sdk.Services().Compute().V1().Instance().Create(ctx, req)
	if err != nil {
		state.Put("error", fmt.Errorf("failed to create instance: %w", err))
		return multistep.ActionHalt
	}

	instanceID := resp.ResourceID()
	state.Put(stateInstanceID, instanceID)

	ui.Message(fmt.Sprintf("Created operation %s with instance %s", resp.ID(), instanceID))
	ui.Message(fmt.Sprintf("Waiting for finish of operation %s...", resp.ID()))

	if err := common.WaitFinishOperationWithTimeout(ctx, s.sdk, resp.ID(), 10*time.Minute); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Instance %s creation completed", instanceID))
	return multistep.ActionContinue
}

func (s *StepCreateInstance) Cleanup(state multistep.StateBag) {
	instanceID, ok := state.Get(stateInstanceID).(string)
	if !ok || instanceID == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Message(fmt.Sprintf("Deleting instance %s...", instanceID))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	resp, err := s.sdk.Services().Compute().V1().Instance().Delete(
		ctx,
		&computev1.DeleteInstanceRequest{
			Id: instanceID,
		},
	)

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete instance %s: %v", instanceID, err))
		return
	}

	if err := common.WaitFinishOperationWithTimeout(ctx, s.sdk, resp.ID(), 10*time.Minute); err != nil {
		ui.Error(fmt.Sprintf("Failed to wait for delete operation %s to finish: %v", resp.ID(), err))
	}

	ui.Message(fmt.Sprintf("Instance %s deletion completed", instanceID))
}

func (s *StepCreateInstance) BuildUserData() string {
	if s.config.Comm.SSHTemporaryKeyPairName == "" {
		// no need to build cloud-init user data if we're not using a temporary SSH key pair,
		// since the public key will be injected through the instance metadata by the communicator step
		return ""
	}

	publicKey := strings.TrimSpace(string(s.config.Comm.SSH.SSHPublicKey))
	if publicKey == "" {
		return ""
	}

	cloudInitUserData := fmt.Sprintf(`#cloud-config
users:
  - default
  - name: %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - %s packer-temporary-key
`, s.config.Comm.SSH.SSHUsername, publicKey)

	return cloudInitUserData
}
