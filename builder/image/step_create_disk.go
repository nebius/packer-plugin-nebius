package image

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/nebius/gosdk"
	commonv1 "github.com/nebius/gosdk/proto/nebius/common/v1"
	computev1 "github.com/nebius/gosdk/proto/nebius/compute/v1"

	"github.com/hashicorp/packer-plugin-nebius/builder/common"
)

const stateDiskID = "disk_id"

type StepCreateDisk struct {
	sdk    *gosdk.SDK
	config Config
}

func NewStepCreateDisk(sdk *gosdk.SDK, config Config) *StepCreateDisk {
	return &StepCreateDisk{
		sdk:    sdk,
		config: config,
	}
}

func (s *StepCreateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Message("Creating disk...")

	diskType, err := diskTypeToProto(s.config.DiskConfig.Type)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	req := &computev1.CreateDiskRequest{
		Metadata: &commonv1.ResourceMetadata{
			Name: fmt.Sprintf("packer-%s-disk", uuid.New().String()),
			Labels: map[string]string{
				"packer_build": s.config.PackerBuildName,
			},
		},
		Spec: &computev1.DiskSpec{
			Size: &computev1.DiskSpec_SizeGibibytes{
				SizeGibibytes: s.config.DiskConfig.SizeGibibytes,
			},
			Type: diskType,
		},
	}

	if s.config.BaseImageConfig.ID != "" {
		req.Spec.Source = &computev1.DiskSpec_SourceImageId{
			SourceImageId: s.config.BaseImageConfig.ID,
		}
	} else {
		req.Spec.Source = &computev1.DiskSpec_SourceImageFamily{
			SourceImageFamily: &computev1.SourceImageFamily{
				ImageFamily: s.config.BaseImageConfig.Family,
				ParentId:    s.config.BaseImageConfig.ParentID,
			},
		}
	}

	resp, err := s.sdk.Services().Compute().V1().Disk().Create(ctx, req)
	if err != nil {
		state.Put("error", fmt.Errorf("failed to create disk: %w", err))
		return multistep.ActionHalt
	}

	diskID := resp.ResourceID()
	state.Put(stateDiskID, diskID)

	ui.Message(fmt.Sprintf("Created operation %s with disk %s", resp.ID(), diskID))
	ui.Message(fmt.Sprintf("Waiting for finish of operation %s...", resp.ID()))

	if err := common.WaitFinishOperationWithTimeout(ctx, s.sdk, resp.ID(), 5*time.Minute); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Disk %s creation completed", diskID))
	return multistep.ActionContinue
}

func (s *StepCreateDisk) Cleanup(state multistep.StateBag) {
	diskID, ok := state.Get(stateDiskID).(string)
	if !ok || diskID == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Message(fmt.Sprintf("Deleting disk %s...", diskID))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	resp, err := s.sdk.Services().Compute().V1().Disk().Delete(
		ctx,
		&computev1.DeleteDiskRequest{
			Id: diskID,
		},
	)

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete disk %s: %v", diskID, err))
		return
	}

	ui.Message(fmt.Sprintf("Disk %s will be deleted in %s operation.", diskID, resp.ID()))
}

func diskTypeToProto(diskType string) (computev1.DiskSpec_DiskType, error) {
	diskType = strings.ToLower(strings.TrimSpace(diskType))

	switch diskType {
	case "network_ssd", "":
		return computev1.DiskSpec_NETWORK_SSD, nil
	case "network_ssd_non_replicated":
		return computev1.DiskSpec_NETWORK_SSD_NON_REPLICATED, nil
	case "network_ssd_io_m3":
		return computev1.DiskSpec_NETWORK_SSD_IO_M3, nil
	default:
		return computev1.DiskSpec_UNSPECIFIED, fmt.Errorf("invalid disk.type: %s", diskType)
	}
}
