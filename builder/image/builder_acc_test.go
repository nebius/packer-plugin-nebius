// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package image

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/hashicorp/packer-plugin-nebius/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/nebius/gosdk"
	v1 "github.com/nebius/gosdk/proto/nebius/compute/v1"
)

//go:embed test-fixtures/template.pkr.hcl
var imageBuilderTemplate string

var sdk *gosdk.SDK
var parentID, publicKey, privateKey, sa string

// Run with: PACKER_ACC=1 go test -count 1 -v ./builder/image/builder_acc_test.go -timeout=120m
func TestAccImageBuilder(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "nebius_image_builder_basic_test",
		Setup: func() error {
			prepareEnvs()

			var err error
			sdk, err = common.NewSDK(
				t.Context(),
				common.ServiceAccountConfig{
					PrivateKey:  privateKey,
					PublicKeyID: publicKey,
					AccountID:   sa,
				},
				parentID,
				"",
				"",
			)

			if err != nil {
				return fmt.Errorf("error creating sdk: %w", err)
			}

			return nil
		},
		Teardown: func() error { return teardown(t.Context()) },
		Template: imageBuilderTemplate,
		Type:     "nebius-image",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("unable find %s", logfile)
			}
			defer logs.Close()

			logsBytes, err := io.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("unable to read %s", logfile)
			}
			logsString := string(logsBytes)

			expectedPatterns := []string{
				`Disk .* creation completed`,
				`Subnet ID not specified, searching for default network...`,
				`Found default subnet with ID: .*`,
				`Creating temporary ED25519 SSH key for instance...`,
				`Instance .* creation completed`,
				`IP address is .*`,
				`Using SSH communicator to connect: .*`,
				`Connected to SSH!`,
				`TASK \[Ensure that net-tools is installed\]`,
				`Cleaning temporary SSH key from guest...`,
				`Temporary SSH key cleaned`,
				`Instance .* stopped`,
				`Image .* created successfully`,
				`Instance .* deletion completed`,
				`Disk .* will be deleted in .* operation\.`,
				`Build 'nebius-image\.acceptance' finished after .*`,
			}

			for _, pattern := range expectedPatterns {
				re := regexp.MustCompile(pattern)
				matches := re.FindStringSubmatch(logsString)
				if matches == nil {
					t.Fatalf("logs don't contain expected pattern %q", pattern)
				}
			}

			return nil
		},
	}

	acctest.TestPlugin(t, testCase)
}

func prepareEnvs() {
	privateKey = os.Getenv("PKR_VAR_NB_PRIVATE_KEY")
	publicKey = os.Getenv("PKR_VAR_NB_PUB_KEY")
	parentID = os.Getenv("PKR_VAR_NB_PARENT_ID")
	sa = os.Getenv("PKR_VAR_NB_SA")
}

func teardown(ctx context.Context) error {
	resp, err := sdk.Services().Compute().V1().Image().List(ctx, &v1.ListImagesRequest{
		ParentId: parentID,
	})
	if err != nil {
		return fmt.Errorf("error listing images during Teardown: %w", err)
	}

	for _, img := range resp.Items {
		_, err := sdk.Services().Compute().V1().Image().Delete(ctx, &v1.DeleteImageRequest{
			Id: img.GetMetadata().GetId(),
		})
		if err != nil {
			return fmt.Errorf("error deleting image: %w", err)
		}
	}

	return nil
}
