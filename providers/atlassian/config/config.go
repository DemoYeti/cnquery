// Copyright (c) Mondoo, Inc.
// SPDX-License-Identifier: BUSL-1.1

package config

import (
	"go.mondoo.com/cnquery/v11/providers-sdk/v1/plugin"
	"go.mondoo.com/cnquery/v11/providers/atlassian/connection/confluence"
	"go.mondoo.com/cnquery/v11/providers/atlassian/provider"
)

var Config = plugin.Provider{
	Name:    "atlassian",
	ID:      "go.mondoo.com/cnquery/v9/providers/atlassian",
	Version: "11.0.28",
	ConnectionTypes: []string{
		provider.DefaultConnectionType,
		"jira",
		"admin",
		string(confluence.Confluence),
		"scim",
	},
	Connectors: []plugin.Connector{
		{
			Name:  "atlassian",
			Use:   "atlassian",
			Short: "an Atlassian Cloud Jira, Confluence or Bitbucket instance",
			Long: `Use the atlassian provider to query resources within Atlassian Cloud, including Jira, Confluence, and SCIM.

Available Commands:
  admin                     Atlassian administrative instance
  jira                      Jira instance
  confluence                Confluence instance
  scim                      SCIM instance

Examples:
  cnquery shell atlassian admin --admin-token <token>
  cnquery shell atlassian jira --host <host> --user <user> --user-token <token>
  cnquery shell atlassian confluence --host <host> --user <user> --user-token <token>
  cnquery shell atlassian scim <directory-id> --scim-token <token>

If you set the ATLASSIAN_ADMIN_TOKEN environment, you can omit the admin-token flag. If you set the ATLASSIAN_USER,
ATLASSIAN_HOST, and ATLASSIAN_USER_TOKEN environment variables, you can omit the user, host, and user-token flags.

For the SCIM token and the directory-id values: Atlassian provides these values when you set up an identity provider.
`,
			MaxArgs:   2,
			Discovery: []string{},
			Flags: []plugin.Flag{
				{
					Long:    "admin-token",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Atlassian admin API token (used for Atlassian admin)",
				},
				{
					Long:    "host",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Atlassian hostname (e.g. https://example.atlassian.net)",
				},
				{
					Long:    "user",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Atlassian user name (e.g. example@example.com)",
				},
				{
					Long:    "user-token",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Atlassian user API token (used for Jira or Confluence)",
				},
				{
					Long:    "scim-token",
					Type:    plugin.FlagType_String,
					Default: "",
					Desc:    "Atlassian SCIM API token (used for SCIM)",
				},
			},
		},
	},
}
