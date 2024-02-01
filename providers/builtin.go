// Copyright (c) Mondoo, Inc.
// SPDX-License-Identifier: BUSL-1.1
//
// This file is auto-generated by 'make providers/config'
// and configured via 'providers.yaml'

package providers

// This is primarily useful for debugging purposes, if you want to
// trace into any provider without having to debug the plugin
// connection separately.

import (
	_ "embed"

	coreconf "go.mondoo.com/cnquery/v10/providers/core/config"
	core "go.mondoo.com/cnquery/v10/providers/core/provider"
)

//go:embed core/resources/core.resources.json
var coreInfo []byte

var builtinProviders = map[string]*builtinProvider{
	coreconf.Config.ID: {
		Runtime: &RunningProvider{
			Name:     coreconf.Config.Name,
			ID:       coreconf.Config.ID,
			Plugin:   core.Init(),
			Schema:   MustLoadSchema("core", coreInfo),
			isClosed: false,
		},
		Config: &coreconf.Config,
	},
	mockProvider.ID: {
		Runtime: &RunningProvider{
			Name:     mockProvider.Name,
			ID:       mockProvider.ID,
			Plugin:   &mockProviderService{coordinator: &GlobalCoordinator},
			isClosed: false,
		},
		Config: mockProvider.Provider,
	},
}
