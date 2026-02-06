package instance

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"golang.org/x/crypto/ssh"
)

const stateSSHPrivateKey = "ssh_private_key"
const stateSSHPublicKey = "ssh_public_key"

type StepCreateSSHKey struct{}

func NewStepCreateSSHKey() *StepCreateSSHKey {
	return &StepCreateSSHKey{}
}

func (s *StepCreateSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Message("Creating SSH key pair...")

	// Generate ed25519 keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		state.Put("error", fmt.Errorf("failed to generate ed25519 key: %w", err))
		return multistep.ActionHalt
	}

	// Convert public key to OpenSSH authorized_keys format:
	// "ssh-ed25519 AAAAC3... comment"
	sshPubKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		state.Put("error", fmt.Errorf("failed to convert public key to ssh format: %w", err))
		return multistep.ActionHalt
	}
	authorizedKeyLine := string(ssh.MarshalAuthorizedKey(sshPubKey))

	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		state.Put("error", fmt.Errorf("failed to marshal private key (pkcs8): %w", err))
		return multistep.ActionHalt
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8,
	})

	state.Put(stateSSHPublicKey, authorizedKeyLine)
	state.Put(stateSSHPrivateKey, string(privatePEM))

	ui.Message("SSH key pair created")
	return multistep.ActionContinue
}

func (s *StepCreateSSHKey) Cleanup(_ multistep.StateBag) {}
