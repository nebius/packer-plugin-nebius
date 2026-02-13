package image

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-nebius/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/nebius/gosdk"
	computev1 "github.com/nebius/gosdk/proto/nebius/compute/v1"
)

type StepStopInstance struct {
	sdk *gosdk.SDK
}

func NewStepStopInstance(sdk *gosdk.SDK) *StepStopInstance {
	return &StepStopInstance{
		sdk: sdk,
	}
}

func (s *StepStopInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Message("Stop instance...")

	resp, err := s.sdk.Services().Compute().V1().Instance().Stop(ctx, &computev1.StopInstanceRequest{
		Id: state.Get(stateInstanceID).(string),
	})
	if err != nil {
		state.Put("error", fmt.Errorf("failed to create instance: %w", err))
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Created operation %s for stopping instance %s", resp.ID(), resp.ResourceID()))
	ui.Message(fmt.Sprintf("Waiting for finish of operation %s...", resp.ID()))

	if err := common.WaitFinishOperationWithTimeout(ctx, s.sdk, resp.ID(), 10*time.Minute); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Instance %s stopped", resp.ResourceID()))
	return multistep.ActionContinue
}

func (s *StepStopInstance) Cleanup(_ multistep.StateBag) {}
