# Copyright IBM Corp. 2020, 2025
# SPDX-License-Identifier: MPL-2.0

locals {
  foo = data.nebius-my-datasource.mock-data.foo
  bar = data.nebius-my-datasource.mock-data.bar
}