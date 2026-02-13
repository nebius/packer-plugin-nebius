#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
env_file="${ACC_ENV_FILE:-${script_dir}/envs.sh}"
if [[ -f "$env_file" ]]; then
  # shellcheck disable=SC1090
  source "$env_file"
fi

required=(
  PKR_VAR_nb_private_key_file
  PKR_VAR_nb_public_key_id
  PKR_VAR_nb_sa_id
  PKR_VAR_nb_parent_id
  PKR_VAR_nb_base_image_family
  PKR_VAR_nb_platform
  PKR_VAR_nb_preset
  PKR_VAR_nb_image_name
  PKR_VAR_nb_image_version
  PKR_VAR_nb_image_family
  PKR_VAR_nb_image_family_human_readable
  PKR_VAR_nb_cpu_architecture
)

missing=()
for var in "${required[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    missing+=("$var")
  fi
done

if (( ${#missing[@]} > 0 )); then
  echo "Missing required env vars:" >&2
  printf '  %s\n' "${missing[@]}" >&2
  exit 1
fi

: "${PKR_VAR_nb_api_endpoint:=}"
export PACKER_ACC=1

cd "${repo_root}"
go test -count 1 -v ./builder/image/builder_acc_test.go -timeout=120m
