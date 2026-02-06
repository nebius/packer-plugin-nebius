package instance

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	nebiuscommon "github.com/hashicorp/packer-plugin-nebius/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/nebius/gosdk"
)

const BuilderId = "nebius.builder"

type Builder struct {
	config Config
	runner multistep.Runner
	sdk    *gosdk.SDK
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	if err := config.Decode(&b.config, &config.DecodeOpts{
		PluginType: BuilderId,
	}, raws...); err != nil {
		return nil, nil, err
	}

	if err := b.config.validate(); err != nil {
		return nil, nil, err
	}

	b.sdk, err = nebiuscommon.NewSDK(context.Background(), b.config.ServiceAccountConfig, b.config.ParentID)
	if err != nil {
		return nil, nil, err
	}

	return []string{}, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	ui.Message("Create resources...")
	steps := []multistep.Step{
		NewStepCreateDisk(b.sdk, b.config),
		NewStepFindNetwork(b.sdk, b.config),
		NewStepCreateSSHKey(),
		NewStepCreateInstance(b.sdk, b.config),
		new(commonsteps.StepProvision),
	}

	// set up the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Set the value of the generated data that will become available to provisioners.
	// To share the data with post-processors, use the StateData in the artifact.
	state.Put("generated_data", map[string]interface{}{
		"GeneratedMockData": "mock-build-data",
	})

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	artifact := &Artifact{
		// Add the builder generated data to the artifact StateData so that post-processors
		// can access them.
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	return artifact, nil
}
