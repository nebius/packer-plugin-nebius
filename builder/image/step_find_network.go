package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/nebius/gosdk"
	v1 "github.com/nebius/gosdk/proto/nebius/vpc/v1"
)

const stateSubnetID = "subnet_id"

type StepFindNetwork struct {
	sdk    *gosdk.SDK
	config Config
}

func NewStepFindNetwork(sdk *gosdk.SDK, config Config) *StepFindNetwork {
	return &StepFindNetwork{
		sdk:    sdk,
		config: config,
	}
}

func (s *StepFindNetwork) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	if s.config.NetworkConfig.SubnetID != "" {
		state.Put(stateSubnetID, s.config.NetworkConfig.SubnetID)
		return multistep.ActionContinue
	}

	ui.Message("Subnet ID not specified, searching for default network...")

	//try to find default network
	getNetworkResp, err := s.sdk.Services().VPC().V1().Network().GetByName(ctx, &v1.GetNetworkByNameRequest{
		Name: "default",
	})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	networkID := getNetworkResp.GetMetadata().GetId()

	// try to find default subnet in the default network
	getSubnetsResp, err := s.sdk.Services().VPC().V1().Subnet().ListByNetwork(ctx, &v1.ListSubnetsByNetworkRequest{
		NetworkId: networkID,
	})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	subnetID := ""
	for _, subnet := range getSubnetsResp.GetItems() {
		if strings.HasPrefix(subnet.GetMetadata().GetName(), "default-subnet-") {
			subnetID = subnet.GetMetadata().GetId()
			break
		}
	}

	if subnetID == "" {
		state.Put("error", fmt.Errorf("no default subnet found in default network"))
		return multistep.ActionHalt
	}

	state.Put(stateSubnetID, subnetID)

	ui.Message(fmt.Sprintf("Found default subnet with ID: %s", subnetID))
	return multistep.ActionContinue
}

func (s *StepFindNetwork) Cleanup(_ multistep.StateBag) {}
