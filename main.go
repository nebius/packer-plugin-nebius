// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	"github.com/hashicorp/packer-plugin-nebius/builder/image"
	scaffoldingVersion "github.com/hashicorp/packer-plugin-nebius/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("image", new(image.Builder))
	pps.SetVersion(scaffoldingVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
