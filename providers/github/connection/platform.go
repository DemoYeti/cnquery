// Copyright (c) Mondoo, Inc.
// SPDX-License-Identifier: BUSL-1.1

package connection

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v57/github"
	"go.mondoo.com/cnquery/v10/providers-sdk/v1/inventory"
)

const (
	DiscoveryAll          = "all"
	DiscoveryAuto         = "auto"
	DiscoveryRepos        = "repos"
	DiscoveryUsers        = "users"
	DiscoveryRepository   = "repository" // deprecated: use repos
	DiscoveryUser         = "user"       // deprecated: use users
	DiscoveryOrganization = "organization"
)

var (
	GithubRepoPlatform = &inventory.Platform{
		Name:    "github-repo",
		Title:   "GitHub Repository",
		Family:  []string{"github"},
		Kind:    "api",
		Runtime: "github",
	}
	GithubUserPlatform = &inventory.Platform{
		Name:    "github-user",
		Title:   "GitHub User",
		Family:  []string{"github"},
		Kind:    "api",
		Runtime: "github",
	}
	GithubOrgPlatform = &inventory.Platform{
		Name:    "github-org",
		Title:   "GitHub Organization",
		Family:  []string{"github"},
		Kind:    "api",
		Runtime: "github",
	}
)

func (c *GithubConnection) PlatformInfo() (*inventory.Platform, error) {
	conf := c.asset.Connections[0]
	if orgId := conf.Options["organization"]; orgId != "" {
		return GithubOrgPlatform, nil
	}

	if userId := conf.Options["user"]; userId != "" {
		return GithubUserPlatform, nil
	}

	_, err := c.Repository()
	if err == nil {
		return GithubRepoPlatform, nil
	}

	return nil, errors.Wrap(err, "could not detect GitHub asset type")
}

func NewGithubOrgIdentifier(orgId string) string {
	return "//platformid.api.mondoo.app/runtime/github/organization/" + orgId
}

func NewGithubUserIdentifier(userId string) string {
	return "//platformid.api.mondoo.app/runtime/github/user/" + userId
}

func NewGitHubRepoIdentifier(ownerId string, repoId string) string {
	return "//platformid.api.mondoo.app/runtime/github/owner/" + ownerId + "/repository/" + repoId
}

func (c *GithubConnection) Identifier() (string, error) {
	conf := c.asset.Connections[0]
	orgId := conf.Options["organization"]
	if orgId != "" {
		return NewGithubOrgIdentifier(orgId), nil
	}

	userId := conf.Options["user"]
	if userId != "" {
		return NewGithubUserIdentifier(userId), nil
	}

	repoId := conf.Options["repository"]
	if repoId != "" {
		ownerId := conf.Options["owner"]
		if ownerId == "" {
			ownerId = conf.Options["organization"]
		}
		if ownerId == "" {
			ownerId = conf.Options["user"]
		}
		return NewGitHubRepoIdentifier(ownerId, repoId), nil
	}

	return "", errors.New("could not identifier GitHub asset")
}

func (c *GithubConnection) Organization() (*github.Organization, error) {
	conf := c.asset.Connections[0]
	orgId := conf.Options["organization"]
	if orgId == "" {
		orgId = conf.Options["owner"]
	}
	if orgId != "" {
		org, _, err := c.Client().Organizations.Get(context.Background(), orgId)
		return org, err
	}

	return nil, errors.New("no organization provided")
}

func (c *GithubConnection) User() (*github.User, error) {
	conf := c.asset.Connections[0]
	userId := conf.Options["user"]
	if userId == "" {
		userId = conf.Options["owner"]
	}

	if userId != "" {
		user, _, err := c.Client().Users.Get(context.Background(), userId)
		return user, err
	}
	return nil, errors.New("no user provided")
}

func (c *GithubConnection) Repository() (*github.Repository, error) {
	conf := c.asset.Connections[0]
	ownerId := conf.Options["owner"]
	if ownerId == "" {
		ownerId = conf.Options["organization"]
	}
	if ownerId == "" {
		ownerId = conf.Options["user"]
	}

	repoId := conf.Options["repository"]
	if ownerId != "" && repoId != "" {
		repo, _, err := c.Client().Repositories.Get(context.Background(), ownerId, repoId)
		return repo, err
	}
	return nil, errors.New("no repository provided")
}
