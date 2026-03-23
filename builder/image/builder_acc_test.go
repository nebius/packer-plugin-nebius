// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package image

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/template.pkr.hcl
var testBuilderHCL2Basic string

var requiredAccEnv = []string{
	"PKR_VAR_NB_PARENT_ID",
	"PKR_VAR_NB_PUB_KEY",
	"PKR_VAR_NB_SA",
	"PKR_VAR_NB_PRIVATE_KEY",
}

func requireAccEnv() error {
	missing := make([]string, 0, len(requiredAccEnv))
	for _, key := range requiredAccEnv {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}

// Run with: PACKER_ACC=1 go test -count 1 -v ./builder/image/builder_acc_test.go -timeout=120m
func TestAccImageBuilder(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "nebius_image_builder_basic_test",
		Setup: func() error {
			return requireAccEnv()
		},
		Teardown: func() error {
			return nil
		},
		Template: testBuilderHCL2Basic,
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
