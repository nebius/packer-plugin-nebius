package image

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/packer-plugin-nebius/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/nebius/gosdk"
	commonv1 "github.com/nebius/gosdk/proto/nebius/common/v1"
	computev1 "github.com/nebius/gosdk/proto/nebius/compute/v1"
)

const stateSecondaryDiskID = "secondary_disk_id"

type StepCreateSecondaryDisk struct {
	sdk    *gosdk.SDK
	config Config
}

func NewStepCreateSecondaryDisk(sdk *gosdk.SDK, config Config) *StepCreateSecondaryDisk {
	return &StepCreateSecondaryDisk{
		sdk:    sdk,
		config: config,
	}
}

func (s *StepCreateSecondaryDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.config.UseSecondaryDisk {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Message("Creating secondary disk...")

	diskType, err := diskTypeToProto(s.config.SecondaryDiskConfig.Type)
	if err != nil {
		state.Put("error", fmt.Errorf("invalid secondary disk type: %w", err))
		return multistep.ActionHalt
	}

	req := &computev1.CreateDiskRequest{
		Metadata: &commonv1.ResourceMetadata{
			Name: fmt.Sprintf("packer-%s-secondary-disk", uuid.New().String()),
			Labels: map[string]string{
				"packer_build": s.config.PackerBuildName,
			},
		},
		Spec: &computev1.DiskSpec{
			Size: &computev1.DiskSpec_SizeGibibytes{
				SizeGibibytes: s.config.SecondaryDiskConfig.SizeGibibytes,
			},
			Type: diskType,
		},
	}

	resp, err := s.sdk.Services().Compute().V1().Disk().Create(ctx, req)
	if err != nil {
		state.Put("error", fmt.Errorf("failed to create secondary disk: %w", err))
		return multistep.ActionHalt
	}

	secondaryDiskID := resp.ResourceID()
	state.Put(stateSecondaryDiskID, secondaryDiskID)

	ui.Message(fmt.Sprintf("Created operation %s with secondary disk %s", resp.ID(), secondaryDiskID))
	ui.Message(fmt.Sprintf("Waiting for finish of operation %s...", resp.ID()))

	if err := common.WaitFinishOperationWithTimeout(ctx, s.sdk, resp.ID(), 5*time.Minute); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Secondary disk %s creation completed", secondaryDiskID))
	return multistep.ActionContinue
}

func (s *StepCreateSecondaryDisk) Cleanup(state multistep.StateBag) {
	secondaryDiskID, ok := state.Get(stateSecondaryDiskID).(string)
	if !ok || secondaryDiskID == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Message(fmt.Sprintf("Deleting secondary disk %s...", secondaryDiskID))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	resp, err := s.sdk.Services().Compute().V1().Disk().Delete(
		ctx,
		&computev1.DeleteDiskRequest{
			Id: secondaryDiskID,
		},
	)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete secondary disk %s: %v", secondaryDiskID, err))
		return
	}

	ui.Message(fmt.Sprintf("Secondary disk %s will be deleted in %s operation.", secondaryDiskID, resp.ID()))
}
