// Copyright (c) Mondoo, Inc.
// SPDX-License-Identifier: BUSL-1.1

package config

import (
	"go.mondoo.com/cnquery/v10/providers-sdk/v1/plugin"
	"go.mondoo.com/cnquery/v10/providers/azure/connection/azureinstancesnapshot"
	"go.mondoo.com/cnquery/v10/providers/azure/provider"
	"go.mondoo.com/cnquery/v10/providers/azure/resources"
)

var Config = plugin.Provider{
	Name:    "azure",
	ID:      "go.mondoo.com/cnquery/v9/providers/azure",
	Version: "10.3.3",
	ConnectionTypes: []string{
		provider.ConnectionType,
		string(azureinstancesnapshot.SnapshotConnectionType),
	},
	Connectors: []plugin.Connector{
		{
			Name:    "azure",
			Use:     "azure",
			Short:   "an Azure subscription",
			MinArgs: 0,
			MaxArgs: 8,
			Discovery: []string{
				resources.DiscoveryAuto,
				resources.DiscoveryAll,
				resources.DiscoverySubscriptions,
				resources.DiscoveryInstances,
				resources.DiscoveryInstancesApi,
				resources.DiscoverySqlServers,
				resources.DiscoveryPostgresServers,
				resources.DiscoveryMySqlServers,
				resources.DiscoveryMariaDbServers,
				resources.DiscoveryStorageAccounts,
				resources.DiscoveryStorageContainers,
				resources.DiscoveryKeyVaults,
				resources.DiscoverySecurityGroups,
			},
			Flags: []plugin.Flag{
				{
					Long:    "tenant-id",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Directory (tenant) ID of the service principal.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "client-id",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Application (client) ID of the service principal.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "client-secret",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Secret for application.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "certificate-path",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Path (in PKCS #12/PFX or PEM format) to the authentication certificate.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "certificate-secret",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Passphrase for the authentication certificate file.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "subscription",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "ID of the Azure subscription to scan.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "subscriptions",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Comma-separated list of Azure subscriptions to include.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "subscriptions-exclude",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Comma-separated list of Azure subscriptions to exclude.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "skip-snapshot-cleanup",
					Type:    plugin.FlagType_Bool,
					Default: "",
					Desc:    "If set, no cleanup will be performed for the snapshot connection.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "skip-snapshot-setup",
					Type:    plugin.FlagType_Bool,
					Default: "",
					Desc:    "If set, no setup will be performed for the snapshot connection. It is expected that the target's disk is already attached. Use together with --lun.",
					Option:  plugin.FlagOption_Hidden,
				},
				{
					Long:    "lun",
					Type:    plugin.FlagType_Int,
					Default: "",
					Desc:    "The logical unit number of the attached disk that should be scanned. Use together with --skip-snapshot-setup.",
					Option:  plugin.FlagOption_Hidden,
				},
			},
		},
	},
}
