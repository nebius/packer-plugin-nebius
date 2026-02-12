package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/nebius/gosdk"
	v1 "github.com/nebius/gosdk/proto/nebius/compute/v1"
)

const (
	stateBaseImageArch        = "base_image_arch"
	stateBaseImageID          = "base_image_id"
	stateBaseImageImageFamily = "base_image_image_family"
	stateBaseImageVersion     = "base_image_version"
)

type StepGetBaseImage struct {
	sdk    *gosdk.SDK
	config Config
}

func NewStepGetBaseImage(sdk *gosdk.SDK, config Config) *StepGetBaseImage {
	return &StepGetBaseImage{
		sdk:    sdk,
		config: config,
	}
}

func (s *StepGetBaseImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Message("Get Base Image...")

	baseImageID := s.config.BaseImageConfig.ID
	if baseImageID == "" {
		disk, err := s.sdk.Services().Compute().V1().Disk().Get(ctx, &v1.GetDiskRequest{
			Id: state.Get(stateDiskID).(string),
		})
		if err != nil {
			state.Put("error", fmt.Errorf("failed to get disk: %w", err))
			return multistep.ActionHalt
		}
		baseImageID = disk.GetStatus().GetSourceImageId()
	}

	var image *v1.Image
	var err error
	image, err = s.sdk.Services().Compute().V1().Image().Get(ctx, &v1.GetImageRequest{
		Id: baseImageID,
	})
	if err != nil {
		state.Put("error", fmt.Errorf("failed to get base image by ID: %w", err))
		return multistep.ActionHalt
	}

	state.Put(stateBaseImageID, image.GetMetadata().GetId())
	state.Put(stateBaseImageImageFamily, image.GetSpec().GetImageFamily())
	state.Put(stateBaseImageVersion, image.GetSpec().GetVersion())
	state.Put(stateBaseImageArch, image.GetSpec().GetCpuArchitecture())

	ui.Message("Base image find")
	return multistep.ActionContinue
}

func (s *StepGetBaseImage) Cleanup(_ multistep.StateBag) {}
