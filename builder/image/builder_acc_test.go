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
	"PKR_VAR_nb_private_key_file",
	"PKR_VAR_nb_public_key_id",
	"PKR_VAR_nb_sa_id",
	"PKR_VAR_nb_parent_id",
	"PKR_VAR_nb_base_image_family",
	"PKR_VAR_nb_platform",
	"PKR_VAR_nb_preset",
	"PKR_VAR_nb_image_name",
	"PKR_VAR_nb_image_version",
	"PKR_VAR_nb_image_family",
	"PKR_VAR_nb_image_family_human_readable",
	"PKR_VAR_nb_cpu_architecture",
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

			expectedLog := "nebius-image.acceptance: Image creation..."
			if matched, _ := regexp.MatchString(expectedLog+".*", logsString); !matched {
				t.Fatalf("logs don't contain expected image creation log: %q", logsString)
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
