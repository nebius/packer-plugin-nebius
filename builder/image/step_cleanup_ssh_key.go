package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepCleanupSSHKey removes the temporary SSH key from the guest so it is not baked into the image.
type StepCleanupSSHKey struct {
	config *Config
}

func NewStepCleanupSSHKey(config *Config) *StepCleanupSSHKey {
	return &StepCleanupSSHKey{config: config}
}

func (s *StepCleanupSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.config.Comm.SSHTemporaryKeyPairName == "" {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packersdk.Ui)

	rawComm, ok := state.GetOk("communicator") // from StepConnect
	if !ok {
		err := fmt.Errorf("communicator missing from state during cleanup")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	comm, ok := rawComm.(packersdk.Communicator)
	if !ok {
		err := fmt.Errorf("communicator value has unexpected type during cleanup")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	username := s.config.Comm.SSH.SSHUsername
	if username == "" {
		err := fmt.Errorf("ssh_username must be set to clean the temporary key")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message("Cleaning temporary SSH key from guest...")

	escaped := escapeShellArg(username)
	scriptBody := fmt.Sprintf(`
set -euo pipefail
HOME_DIR=$(getent passwd %s | cut -d: -f6)
if [ -n "$HOME_DIR" ]; then
  AUTH_FILE="$HOME_DIR/.ssh/authorized_keys"
  if [ -f "$AUTH_FILE" ]; then
    sed -i "/%s/d" "$AUTH_FILE" || true
  fi
fi
cloud-init clean --logs >/dev/null 2>&1 || true
`, escaped, escapeSedPattern(s.config.Comm.SSH.SSHTemporaryKeyPairName))

	cmd := &packersdk.RemoteCmd{
		Command: fmt.Sprintf("sudo bash -c %s", escapeShellArg(scriptBody)),
	}

	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		ui.Error(fmt.Sprintf("Failed to remove temporary SSH key: %s", err))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message("Temporary SSH key cleaned")
	return multistep.ActionContinue
}

func (s *StepCleanupSSHKey) Cleanup(_ multistep.StateBag) {}

func escapeShellArg(arg string) string {
	if arg == "" {
		return ""
	}
	return "'" + strings.ReplaceAll(arg, "'", `'"'"'`) + "'"
}

func escapeSedPattern(input string) string {
	if input == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"/", "\\/",
		".", "\\.",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"^", "\\^",
		"$", "\\$",
	)
	return replacer.Replace(input)
}
