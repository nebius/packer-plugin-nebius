# Copy to envs.ssh and fill in real values.
# This file is sourced by scripts/run-acc-image-test.sh.

export PKR_VAR_nb_private_key_file="/absolute/path/to/private.pem"
export PKR_VAR_nb_public_key_id="publickey-xxxxxxxxxxxxxxxx"
export PKR_VAR_nb_sa_id="serviceaccount-xxxxxxxxxxxxxxxx"
export PKR_VAR_nb_parent_id="project-xxxxxxxxxxxxxxxx"
export PKR_VAR_nb_base_image_family="ubuntu24.04-driverless"
export PKR_VAR_nb_platform="cpu-d3"
export PKR_VAR_nb_preset="4vcpu-16gb"
export PKR_VAR_nb_image_name="ubuntu24.04-driverless-0.0.9"
export PKR_VAR_nb_image_version="0.0.9"
export PKR_VAR_nb_image_family="ubuntu24.04-driverless"
export PKR_VAR_nb_image_family_human_readable="Ubuntu 24.04 Driverless"
export PKR_VAR_nb_cpu_architecture="amd64"

# Optional
# export PKR_VAR_nb_api_endpoint="api.testing.nebius.cloud:443"
