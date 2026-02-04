// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	"github.com/hashicorp/packer-plugin-nebius/builder/instance"
	scaffoldingData "github.com/hashicorp/packer-plugin-nebius/datasource/scaffolding"
	scaffoldingPP "github.com/hashicorp/packer-plugin-nebius/post-processor/scaffolding"
	scaffoldingProv "github.com/hashicorp/packer-plugin-nebius/provisioner/scaffolding"
	scaffoldingVersion "github.com/hashicorp/packer-plugin-nebius/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("instance", new(instance.Builder))
	pps.RegisterProvisioner("my-provisioner", new(scaffoldingProv.Provisioner))
	pps.RegisterPostProcessor("my-post-processor", new(scaffoldingPP.PostProcessor))
	pps.RegisterDatasource("my-datasource", new(scaffoldingData.Datasource))
	pps.SetVersion(scaffoldingVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
