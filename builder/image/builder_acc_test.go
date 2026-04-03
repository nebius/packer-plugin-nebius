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

//go:embed test-fixtures/boot_disk_image_template.pkr.hcl
var bootDiskImageBuilderTemplate string

//go:embed test-fixtures/secondary_disk_image_template.pkr.hcl
var secondaryDiskImageBuilderTemplate string

var sdk *gosdk.SDK
var parentID, publicKey, privateKey, sa string

// Run with: PACKER_ACC=1 go test -count 1 -v ./builder/image/builder_acc_test.go -timeout=120m
func TestAccImageBuilder(t *testing.T) {
	t.Run("boot disk image", func(t *testing.T) {
		runAcceptanceTest(t, acceptanceTestCase{
			name:     "nebius_boot_disk_image_builder_test",
			template: bootDiskImageBuilderTemplate,
			expectedPatterns: []string{
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
				`Build 'nebius-image\.acceptance-boot-disk-image' finished after .*`,
			},
			imageCheck: assertBootDiskImage,
		})
	})

	t.Run("secondary disk image", func(t *testing.T) {
		runAcceptanceTest(t, acceptanceTestCase{
			name:     "nebius_secondary_disk_image_builder_test",
			template: secondaryDiskImageBuilderTemplate,
			expectedPatterns: []string{
				`Disk .* creation completed`,
				`Subnet ID not specified, searching for default network...`,
				`Found default subnet with ID: .*`,
				`Creating secondary disk...`,
				`Secondary disk .* creation completed`,
				`Creating temporary ED25519 SSH key for instance...`,
				`Instance .* creation completed`,
				`IP address is .*`,
				`Using SSH communicator to connect: .*`,
				`Connected to SSH!`,
				`TASK \[Format secondary disk\]`,
				`Cleaning temporary SSH key from guest...`,
				`Temporary SSH key cleaned`,
				`Instance .* stopped`,
				`Image .* created successfully`,
				`Instance .* deletion completed`,
				`Secondary disk .* will be deleted in .* operation\.`,
				`Disk .* will be deleted in .* operation\.`,
				`Build 'nebius-image\.acceptance-secondary-disk-image' finished after .*`,
			},
			imageCheck: assertSecondaryDiskImage,
		})
	})
}

type acceptanceTestCase struct {
	name             string
	template         string
	expectedPatterns []string
	imageCheck       func(context.Context, string) error
}

func runAcceptanceTest(t *testing.T, tc acceptanceTestCase) {
	t.Helper()

	testCase := &acctest.PluginTestCase{
		Name:     tc.name,
		Setup:    func() error { return setupSDK(t.Context()) },
		Teardown: func() error { return teardown(t.Context()) },
		Template: tc.template,
		Type:     "nebius-image",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			logsString, err := readBuildLogs(buildCommand, logfile)
			if err != nil {
				return err
			}

			for _, pattern := range tc.expectedPatterns {
				re := regexp.MustCompile(pattern)
				matches := re.FindStringSubmatch(logsString)
				if matches == nil {
					t.Fatalf("logs don't contain expected pattern %q", pattern)
				}
			}

			if tc.imageCheck != nil {
				imageID, err := extractCreatedImageID(logsString)
				if err != nil {
					return err
				}

				if err := tc.imageCheck(t.Context(), imageID); err != nil {
					return err
				}
			}

			return nil
		},
	}

	acctest.TestPlugin(t, testCase)
}

func setupSDK(ctx context.Context) error {
	prepareEnvs()

	var err error
	sdk, err = common.NewSDK(
		ctx,
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
}

func readBuildLogs(buildCommand *exec.Cmd, logfile string) (string, error) {
	if buildCommand.ProcessState != nil && buildCommand.ProcessState.ExitCode() != 0 {
		return "", fmt.Errorf("bad exit code. Logfile: %s", logfile)
	}

	logs, err := os.Open(logfile)
	if err != nil {
		return "", fmt.Errorf("unable find %s", logfile)
	}
	defer logs.Close()

	logsBytes, err := io.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("unable to read %s", logfile)
	}

	return string(logsBytes), nil
}

func extractCreatedImageID(logs string) (string, error) {
	matches := regexp.MustCompile(`Image ([^ ]+) created successfully`).FindStringSubmatch(logs)
	if len(matches) != 2 {
		return "", fmt.Errorf("unable to find created image ID in logs")
	}

	return matches[1], nil
}

func assertBootDiskImage(ctx context.Context, imageID string) error {
	image, err := sdk.Services().Compute().V1().Image().Get(ctx, &v1.GetImageRequest{
		Id: imageID,
	})
	if err != nil {
		return fmt.Errorf("error getting image %s: %w", imageID, err)
	}

	labels := image.GetMetadata().GetLabels()
	for _, key := range []string{"based_on_image_family", "based_on_image_version"} {
		if labels[key] == "" {
			return fmt.Errorf("expected %q label on boot disk image", key)
		}
	}

	if _, ok := labels["based_on_image_id"]; ok {
		return fmt.Errorf("unexpected %q label on boot disk image", "based_on_image_id")
	}

	return nil
}

func assertSecondaryDiskImage(ctx context.Context, imageID string) error {
	image, err := sdk.Services().Compute().V1().Image().Get(ctx, &v1.GetImageRequest{
		Id: imageID,
	})
	if err != nil {
		return fmt.Errorf("error getting image %s: %w", imageID, err)
	}

	labels := image.GetMetadata().GetLabels()
	for _, key := range []string{"based_on_image_id", "based_on_image_family", "based_on_image_version"} {
		if _, ok := labels[key]; ok {
			return fmt.Errorf("unexpected %q label on secondary disk image", key)
		}
	}

	return nil
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
