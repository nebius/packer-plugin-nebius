package image

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/nebius/gosdk"
	computev1 "github.com/nebius/gosdk/proto/nebius/compute/v1"
)

type StepGetInstanceIP struct {
	sdk    *gosdk.SDK
	config Config
}

func NewStepGetInstanceIP(sdk *gosdk.SDK, config Config) *StepGetInstanceIP {
	return &StepGetInstanceIP{
		sdk:    sdk,
		config: config,
	}
}

func (s *StepGetInstanceIP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Message(fmt.Sprintf("Get instance ip address..."))

	instance, err := s.sdk.Services().Compute().V1().Instance().Get(ctx, &computev1.GetInstanceRequest{
		Id: state.Get(stateInstanceID).(string),
	})
	if err != nil {
		state.Put("error", fmt.Errorf("failed to get instance details: %w", err))
		return multistep.ActionHalt
	}

	if len(instance.GetStatus().GetNetworkInterfaces()) == 0 {
		state.Put("error", fmt.Errorf("instance has no network interfaces"))
		return multistep.ActionHalt
	}

	ipAddress := ""
	if s.config.NetworkConfig.AssociatePublicIpAddress {
		ipAddress = instance.GetStatus().GetNetworkInterfaces()[0].GetPublicIpAddress().GetAddress()
	} else {
		ui.Message("associate_public_ip_address is not set, using private IP address")
		ipAddress = instance.GetStatus().GetNetworkInterfaces()[0].GetIpAddress().GetAddress()
	}

	cleanIP, err := ensurePlainIP(ipAddress)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("IP address is %s", cleanIP))
	state.Put(stateIPAddress, cleanIP)

	return multistep.ActionContinue
}

func (s *StepGetInstanceIP) Cleanup(_ multistep.StateBag) {}

func ensurePlainIP(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("instance IP address is empty")
	}

	if strings.Contains(raw, "/") {
		raw = strings.SplitN(raw, "/", 2)[0]
	}

	if ip := net.ParseIP(raw); ip == nil {
		return "", fmt.Errorf("instance IP address %q is not valid", raw)
	}

	return raw, nil
}
