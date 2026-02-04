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

const stateDiskID = "disk_id"

type StepCreateDisk struct {
	SDK    *gosdk.SDK
	Config Config
}

func (s *StepCreateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Message("Creating disk...")

	diskType, err := diskTypeToProto(s.Config.DiskConfig.Type)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	req := &computev1.CreateDiskRequest{
		Metadata: &commonv1.ResourceMetadata{
			Name: uuid.New().String(),
		},
		Spec: &computev1.DiskSpec{
			Size: &computev1.DiskSpec_SizeGibibytes{
				SizeGibibytes: s.Config.DiskConfig.SizeGibibytes,
			},
			Type: diskType,
		},
	}

	if s.Config.BaseImageConfig.ID != "" {
		req.Spec.Source = &computev1.DiskSpec_SourceImageId{
			SourceImageId: s.Config.BaseImageConfig.ID,
		}
	} else {
		req.Spec.Source = &computev1.DiskSpec_SourceImageFamily{
			SourceImageFamily: &computev1.SourceImageFamily{
				ImageFamily: s.Config.BaseImageConfig.Family,
				ParentId:    s.Config.ParentID,
			},
		}
	}

	resp, err := s.SDK.Services().Compute().V1().Disk().Create(ctx, req)
	if err != nil {
		state.Put("error", fmt.Errorf("failed to create disk: %w", err))
		return multistep.ActionHalt
	}

	diskID := resp.ResourceID()
	state.Put(stateDiskID, diskID)

	ui.Message(fmt.Sprintf("Created operation %s with disk %s", resp.ID(), diskID))
	ui.Message(fmt.Sprintf("Waiting for finish of operation %s...", resp.ID()))

	if err := common.WaitFinishOperation(ctx, s.SDK, resp.ID()); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Disk %s creation complited", diskID))
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

	resp, err := s.SDK.Services().Compute().V1().Disk().Delete(
		ctx,
		&computev1.DeleteDiskRequest{
			Id: diskID,
		},
	)

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete disk %s: %v", diskID, err))
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
