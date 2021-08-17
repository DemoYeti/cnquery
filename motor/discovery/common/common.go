package common

import (
	"github.com/rs/zerolog/log"
	"go.mondoo.io/mondoo/motor/asset"
	"go.mondoo.io/mondoo/motor/vault"
)

type (
	// CredentialFn retrieves the credentials to connect to the platform
	CredentialFn func(cred *vault.Credential) (*vault.Credential, error)
	// QuerySecretFn is used during discovery phase to identify a secret for an asset
	QuerySecretFn func(a *asset.Asset) (*vault.Credential, error)
)

func EnrichAssetWithSecrets(a *asset.Asset, sfn QuerySecretFn) {
	for j := range a.Connections {
		conn := a.Connections[j]

		if len(conn.Credentials) == 0 {
			creds, err := sfn(a)
			if err == nil {
				conn.Credentials = []*vault.Credential{creds}
			} else {
				log.Warn().Str("name", a.Name).Msg("could not determine credentials for asset")
			}
		}
	}
}
