## The Example Folder

This folder showcases a working `nebius-instance` configuration that matches the plugin logic from `builder/instance`.
It defines `required_plugins` and wires up every step of your builder flow so that the GitHub Action can run
`packer init -upgrade .` followed by `packer build .` as part of every PR.

The example configuration (`build.pkr.hcl`) highlights:

- Explicit `service_account`, `disk`, `base_image`, `network`, and `instance` blocks to drive the Nebius APIs.
- An SSH communicator that mirrors the builder’s `communicator` defaults and uses the generated temporary keys.
- The `ansible` provisioner that relies on `StepConnect` to establish the SSH session before running playbooks.

During the build, the plugin will:
1. Generate a temporary SSH keypair with `StepSSHKeyGen`.
2. Create the instance, poll its interfaces for an IP, and store that IP in the state bag.
3. Connect via the communicator configured with that IP and the generated private key.
4. Run `provision.yml`, which simply installs `net-tools` over SSH so you know the system is working.

Authentication secrets (e.g., service account key files, public key IDs) are deliberately absent from source control.
Provide them via environment variables or GitHub Secrets and reference them in
[.github/workflows/test-plugin-example.yml](/.github/workflows/test-plugin-example.yml). Example:

```yml
  - name: Build
    working-directory: ${{ github.event.inputs.folder }}
    run: PACKER_LOG=${{ github.event.inputs.logs }} packer build .
    env:
      NB_AUTHKEY_PRIVATE_PATH: ${{ secrets.NB_AUTHKEY_PRIVATE_PATH }}
      NB_AUTHKEY_PUBLIC_ID: ${{ secrets.NB_AUTHKEY_PUBLIC_ID }}
      NB_SA_ID: ${{ secrets.NB_SA_ID }}
```
