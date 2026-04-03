package image

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer-plugin-nebius/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/nebius/gosdk"
	commonv1 "github.com/nebius/gosdk/proto/nebius/common/v1"
	computev1 "github.com/nebius/gosdk/proto/nebius/compute/v1"
)

const stateImageID = "image_id"

type StepImageCreate struct {
	sdk    *gosdk.SDK
	config Config
}

func NewStepCreateImage(sdk *gosdk.SDK, config Config) *StepImageCreate {
	return &StepImageCreate{
		sdk:    sdk,
		config: config,
	}
}

func (s *StepImageCreate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Message("Image creation...")

	labels := copyLabels(s.config.ImageConfig.Labels)
	diskID := state.Get(stateDiskID).(string)
	cpuArchitecture := getCPUArchitecture(
		s.config.ImageConfig.CPUArchitecture,
		state.Get(stateBaseImageArch).(computev1.ImageSpec_CPUArchitecture),
	)

	if s.config.UseSecondaryDisk {
		diskID = state.Get(stateSecondaryDiskID).(string)
		cpuArchitecture = getCPUArchitecture(s.config.ImageConfig.CPUArchitecture, computev1.ImageSpec_UNSPECIFIED)
	} else if s.config.BaseImageConfig.ID != "" {
		labels["based_on_image_id"] = state.Get(stateBaseImageID).(string)
	} else {
		labels["based_on_image_family"] = state.Get(stateBaseImageImageFamily).(string)
		labels["based_on_image_version"] = state.Get(stateBaseImageVersion).(string)
	}

	req := &computev1.CreateImageRequest{
		Metadata: &commonv1.ResourceMetadata{
			Name:     s.config.ImageConfig.Name,
			Labels:   labels,
			ParentId: s.config.ImageConfig.ParentID,
		},
		Spec: &computev1.ImageSpec{
			ImageFamily:     s.config.ImageConfig.ImageFamily,
			Source:          &computev1.ImageSpec_SourceDiskId{SourceDiskId: diskID},
			CpuArchitecture: cpuArchitecture,
		},
	}

	if req.GetSpec().GetImageFamily() != "" {
		req.GetSpec().Version = s.config.ImageConfig.Version
		req.GetSpec().ImageFamilyHumanReadable = s.config.ImageConfig.ImageFamilyHumanReadable
		req.GetSpec().RecommendedPlatforms = s.config.ImageConfig.RecommendedPlatforms
		req.GetSpec().UnsupportedPlatforms = s.config.ImageConfig.UnsupportedPlatforms
	}

	resp, err := s.sdk.Services().Compute().V1().Image().Create(ctx, req)
	if err != nil {
		state.Put("error", fmt.Errorf("failed to create image: %w", err))
		return multistep.ActionHalt
	}

	opID, imageID := resp.ID(), resp.ResourceID()
	state.Put(stateImageID, imageID)

	ui.Message(fmt.Sprintf("Created operation %s for image %s creation", opID, imageID))
	ui.Message(fmt.Sprintf("Waiting for finish of operation %s...", opID))

	if err := common.WaitFinishOperationWithTimeout(ctx, s.sdk, opID, 10*time.Minute); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Image %s created successfully", imageID))
	return multistep.ActionContinue
}

func (s *StepImageCreate) Cleanup(_ multistep.StateBag) {}

func getCPUArchitecture(providedCPUArch string, baseImageCPUArch computev1.ImageSpec_CPUArchitecture) computev1.ImageSpec_CPUArchitecture {
	if baseImageCPUArch != computev1.ImageSpec_UNSPECIFIED {
		return baseImageCPUArch
	}

	normalizedProvidedCPUArch := strings.ToLower(strings.TrimSpace(providedCPUArch))
	if normalizedProvidedCPUArch == "amd64" {
		return computev1.ImageSpec_AMD64
	} else if normalizedProvidedCPUArch == "arm64" {
		return computev1.ImageSpec_ARM64
	}

	return computev1.ImageSpec_UNSPECIFIED
}

func copyLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(labels))
	for key, value := range labels {
		cloned[key] = value
	}

	return cloned
}
