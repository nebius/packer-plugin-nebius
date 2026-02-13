package image

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hcldec"
	nebiuscommon "github.com/hashicorp/packer-plugin-nebius/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/nebius/gosdk"
)

const BuilderId = "nebius.image-builder"

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

	if errs := b.config.prepareSSH(interpolate.NewContext()); len(errs) > 0 {
		return nil, nil, multierror.Append(nil, errs...).ErrorOrNil()
	}

	if errs := b.config.validate(); len(errs) > 0 {
		return nil, nil, multierror.Append(nil, errs...).ErrorOrNil()
	}

	b.sdk, err = nebiuscommon.NewSDK(context.Background(), b.config.ServiceAccountConfig, b.config.ParentID, b.config.APIEndpoint)
	if err != nil {
		return nil, nil, err
	}

	return []string{}, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	ui.Message("Create resources...")
	steps := []multistep.Step{
		NewStepCreateDisk(b.sdk, b.config),
		NewStepGetBaseImage(b.sdk, b.config),
		NewStepFindNetwork(b.sdk, b.config),
		&communicator.StepSSHKeyGen{
			CommConf:            &b.config.Comm,
			SSHTemporaryKeyPair: b.config.Comm.SSH.SSHTemporaryKeyPair,
		},
		NewStepCreateInstance(b.sdk, &b.config),
		NewStepGetInstanceIP(b.sdk, b.config),
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      communicator.CommHost("", stateIPAddress),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		new(commonsteps.StepProvision), // Ansible, etc. provisioning steps would go here
		NewStepCleanupSSHKey(&b.config),
		NewStepStopInstance(b.sdk),
		NewStepCreateImage(b.sdk, b.config),
	}

	// set up the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	var imageID string
	if raw, ok := state.GetOk(stateImageID); ok {
		if str, isStr := raw.(string); isStr {
			imageID = str
		}
	}

	artifact := &Artifact{
		StateData: map[string]interface{}{
			"image_id": imageID,
		},
		imageID: imageID,
	}
	return artifact, nil
}
