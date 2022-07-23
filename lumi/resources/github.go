package resources

import (
	"context"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v45/github"
	"github.com/rs/zerolog/log"
	"go.mondoo.io/mondoo/lumi"
	"go.mondoo.io/mondoo/motor/providers"
	gh_transport "go.mondoo.io/mondoo/motor/providers/github"
)

func githubtransport(t providers.Transport) (*gh_transport.Provider, error) {
	gt, ok := t.(*gh_transport.Provider)
	if !ok {
		return nil, errors.New("github resource is not supported on this transport")
	}
	return gt, nil
}

func githubTimestamp(ts *github.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	return &ts.Time
}

func (g *lumiGithub) id() (string, error) {
	return "github", nil
}

func (g *lumiGithub) GetUser() (interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	user, err := gt.User()
	if err != nil {
		return nil, err
	}
	var x interface{}
	x = user
	return x, nil
}

func (g *lumiGithub) GetRepositories() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	user, err := gt.User()
	if err != nil {
		return nil, err
	}

	repos, _, err := gt.Client().Repositories.List(context.Background(), user.GetLogin(), &github.RepositoryListOptions{})
	if err != nil {
		return nil, err
	}

	res := []interface{}{}
	for i := range repos {
		repo := repos[i]

		var id int64
		if repo.ID != nil {
			id = *repo.ID
		}

		owner, err := g.MotorRuntime.CreateResource("github.user",
			"id", repo.GetOwner().GetID(),
			"login", repo.GetOwner().GetLogin(),
		)
		if err != nil {
			return nil, err
		}

		r, err := g.MotorRuntime.CreateResource("github.repository",
			"id", id,
			"name", toString(repo.Name),
			"fullName", toString(repo.FullName),
			"description", toString(repo.Description),
			"homepage", toString(repo.Homepage),
			"createdAt", githubTimestamp(repo.CreatedAt),
			"updatedAt", githubTimestamp(repo.UpdatedAt),
			"archived", toBool(repo.Archived),
			"disabled", toBool(repo.Disabled),
			"private", toBool(repo.Private),
			"visibility", toString(repo.Visibility),
			"allowAutoMerge", toBool(repo.AllowAutoMerge),
			"allowForking", toBool(repo.AllowForking),
			"allowMergeCommit", toBool(repo.AllowMergeCommit),
			"allowRebaseMerge", toBool(repo.AllowRebaseMerge),
			"allowSquashMerge", toBool(repo.AllowSquashMerge),
			"hasIssues", toBool(repo.HasIssues),
			"organizationName", "",
			"defaultBranchName", toString(repo.DefaultBranch),
			"owner", owner,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (g *lumiGithubOrganization) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return "github.organization/" + strconv.FormatInt(id, 10), nil
}

func (g *lumiGithubOrganization) init(args *lumi.Args) (*lumi.Args, GithubOrganization, error) {
	if len(*args) > 2 {
		return args, nil, nil
	}

	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, nil, err
	}

	org, err := gt.Organization()
	if err != nil {
		return nil, nil, err
	}

	(*args)["id"] = toInt64(org.ID)
	(*args)["name"] = toString(org.Name)
	(*args)["login"] = toString(org.Login)
	(*args)["nodeId"] = toString(org.NodeID)
	(*args)["company"] = toString(org.Company)
	(*args)["blog"] = toString(org.Blog)
	(*args)["location"] = toString(org.Location)
	(*args)["email"] = toString(org.Email)
	(*args)["twitterUsername"] = toString(org.TwitterUsername)
	(*args)["description"] = toString(org.Description)
	(*args)["createdAt"] = org.CreatedAt
	(*args)["updatedAt"] = org.UpdatedAt
	(*args)["totalPrivateRepos"] = toInt(org.TotalPrivateRepos)
	(*args)["ownedPrivateRepos"] = toInt(org.OwnedPrivateRepos)
	(*args)["privateGists"] = toInt(org.PrivateGists)
	(*args)["diskUsage"] = toInt(org.DiskUsage)
	(*args)["collaborators"] = toInt(org.Collaborators)
	(*args)["billingEmail"] = toString(org.BillingEmail)

	plan, _ := jsonToDict(org.Plan)
	(*args)["plan"] = plan

	(*args)["twoFactorRequirementEnabled"] = toBool(org.TwoFactorRequirementEnabled)
	(*args)["isVerified"] = toBool(org.IsVerified)

	(*args)["defaultRepositoryPermission"] = toString(org.DefaultRepoPermission)
	(*args)["membersCanCreateRepositories"] = toBool(org.MembersCanCreateRepos)
	(*args)["membersCanCreatePublicRepositories"] = toBool(org.MembersCanCreatePublicRepos)
	(*args)["membersCanCreatePrivateRepositories"] = toBool(org.MembersCanCreatePrivateRepos)
	(*args)["membersCanCreateInternalRepositories"] = toBool(org.MembersCanCreateInternalRepos)
	(*args)["membersCanCreatePages"] = toBool(org.MembersCanCreatePages)
	(*args)["membersCanCreatePublicPages"] = toBool(org.MembersCanCreatePublicPages)
	(*args)["membersCanCreatePrivatePages"] = toBool(org.MembersCanCreatePrivateRepos)

	return args, nil, nil
}

func (g *lumiGithubOrganization) GetMembers() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	orgLogin, err := g.Login()
	if err != nil {
		return nil, err
	}
	members, _, err := gt.Client().Organizations.ListMembers(context.Background(), orgLogin, nil)
	if err != nil {
		return nil, err
	}

	res := []interface{}{}
	for i := range members {
		member := members[i]

		r, err := g.MotorRuntime.CreateResource("github.user",
			"id", toInt64(member.ID),
			"login", toString(member.Login),
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (g *lumiGithubOrganization) GetOwners() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	orgLogin, err := g.Login()
	if err != nil {
		return nil, err
	}

	members, _, err := gt.Client().Organizations.ListMembers(context.Background(), orgLogin, &github.ListMembersOptions{
		Role: "admin",
	})
	if err != nil {
		return nil, err
	}

	res := []interface{}{}
	for i := range members {
		member := members[i]

		var id int64
		if member.ID != nil {
			id = *member.ID
		}

		r, err := g.MotorRuntime.CreateResource("github.user",
			"id", id,
			"login", toString(member.Login),
			"name", toString(member.Name),
			"email", toString(member.Email),
			"bio", toString(member.Bio),
			"createdAt", githubTimestamp(member.CreatedAt),
			"updatedAt", githubTimestamp(member.UpdatedAt),
			"suspendedAt", githubTimestamp(member.SuspendedAt),
			"company", toString(member.Company),
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (g *lumiGithubOrganization) GetTeams() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	orgLogin, err := g.Login()
	if err != nil {
		return nil, err
	}
	teams, _, err := gt.Client().Teams.ListTeams(context.Background(), orgLogin, nil)
	if err != nil {
		return nil, err
	}

	res := []interface{}{}
	for i := range teams {
		team := teams[i]
		r, err := g.MotorRuntime.CreateResource("github.team",
			"id", toInt64(team.ID),
			"name", toString(team.Name),
			"description", toString(team.Description),
			"slug", toString(team.Slug),
			"privacy", toString(team.Privacy),
			"defaultPermission", toString(team.Permission),
			"organization", g,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (g *lumiGithubOrganization) GetRepositories() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	orgLogin, err := g.Login()
	if err != nil {
		return nil, err
	}

	repos, _, err := gt.Client().Repositories.ListByOrg(context.Background(), orgLogin, &github.RepositoryListByOrgOptions{Type: "all"})
	if err != nil {
		return nil, err
	}

	res := []interface{}{}
	for i := range repos {
		repo := repos[i]

		var id int64
		if repo.ID != nil {
			id = *repo.ID
		}

		owner, err := g.MotorRuntime.CreateResource("github.user",
			"id", repo.GetOwner().GetID(),
			"login", repo.GetOwner().GetLogin(),
		)
		if err != nil {
			return nil, err
		}

		r, err := g.MotorRuntime.CreateResource("github.repository",
			"id", id,
			"name", toString(repo.Name),
			"fullName", toString(repo.FullName),
			"description", toString(repo.Description),
			"homepage", toString(repo.Homepage),
			"createdAt", githubTimestamp(repo.CreatedAt),
			"updatedAt", githubTimestamp(repo.UpdatedAt),
			"archived", toBool(repo.Archived),
			"disabled", toBool(repo.Disabled),
			"private", toBool(repo.Private),
			"visibility", toString(repo.Visibility),
			"allowAutoMerge", toBool(repo.AllowAutoMerge),
			"allowForking", toBool(repo.AllowForking),
			"allowMergeCommit", toBool(repo.AllowMergeCommit),
			"allowRebaseMerge", toBool(repo.AllowRebaseMerge),
			"allowSquashMerge", toBool(repo.AllowSquashMerge),
			"hasIssues", toBool(repo.HasIssues),
			"organizationName", orgLogin,
			"defaultBranchName", toString(repo.DefaultBranch),
			"owner", owner,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (g *lumiGithubOrganization) GetWebhooks() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	ownerLogin, err := g.Login()
	if err != nil {
		return nil, err
	}

	hooks, _, err := gt.Client().Organizations.ListHooks(context.TODO(), ownerLogin, &github.ListOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range hooks {
		h := hooks[i]
		config, err := jsonToDict(h.Config)
		if err != nil {
			return nil, err
		}

		lumiUser, err := g.MotorRuntime.CreateResource("github.webhook",
			"id", toInt64(h.ID),
			"name", toString(h.Name),
			"events", sliceInterface(h.Events),
			"config", config,
			"url", toString(h.URL),
			"name", toString(h.Name),
			"active", toBool(h.Active),
		)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiUser)
	}

	return res, nil
}

func (g *lumiGithubPackage) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return "github.package/" + strconv.FormatInt(id, 10), nil
}

func (g *lumiGithubOrganization) GetPackages() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	ownerLogin, err := g.Login()
	if err != nil {
		return nil, err
	}

	pkgTypes := []string{"npm", "maven", "rubygems", "docker", "nuget", "container"}
	res := []interface{}{}
	for i := range pkgTypes {
		packages, _, err := gt.Client().Organizations.ListPackages(context.Background(), ownerLogin, &github.PackageListOptions{
			PackageType: github.String(pkgTypes[i]),
		})
		if err != nil {
			log.Error().Err(err).Msg("unable to get hooks list")
			if strings.Contains(err.Error(), "404") {
				return nil, nil
			}
			return nil, err
		}

		for i := range packages {
			p := packages[i]

			owner, err := g.MotorRuntime.CreateResource("github.user",
				"id", p.GetOwner().GetID(),
				"login", p.GetOwner().GetLogin(),
			)
			if err != nil {
				return nil, err
			}

			lumiGhPackage, err := g.MotorRuntime.CreateResource("github.package",
				"id", toInt64(p.ID),
				"name", toString(p.Name),
				"packageType", toString(p.PackageType),
				"owner", owner,
				"createdAt", githubTimestamp(p.CreatedAt),
				"updatedAt", githubTimestamp(p.UpdatedAt),
				"versionCount", toInt64(p.VersionCount),
				"visibility", toString(p.Visibility),
			)
			if err != nil {
				return nil, err
			}
			pkg := lumiGhPackage.(GithubPackage)

			// NOTE: we need to fetch repo separately because the Github repo object is not complete, instead of
			// call the repo fetching all the time, we make this lazy loading
			if p.Repository != nil && p.Repository.Name != nil {
				pkg.LumiResource().Cache.Store("_repository", &lumi.CacheEntry{Data: toString(p.Repository.Name)})
			}
			res = append(res, pkg)
		}
	}

	return res, nil
}

func (g *lumiGithubPackage) GetRepository() (interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	entry, ok := g.Cache.Load("_repository")
	if !ok {
		return nil, errors.New("could not load the repository")
	}

	repoName := entry.Data.(string)

	owner, err := g.Owner()
	if err != nil {
		return nil, err
	}

	ownerLogin, err := owner.Login()
	if err != nil {
		return nil, err
	}

	repo, _, err := gt.Client().Repositories.Get(context.Background(), ownerLogin, repoName)
	if err != nil {
		return nil, err
	}
	return newLumiGithubRepository(g.MotorRuntime, repo)
}

func (g *lumiGithubOrganization) GetInstallations() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	orgLogin, err := g.Login()
	if err != nil {
		return nil, err
	}

	apps, _, err := gt.Client().Organizations.ListInstallations(context.Background(), orgLogin, &github.ListOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}

	res := []interface{}{}
	for i := range apps.Installations {
		app := apps.Installations[i]

		var id int64
		if app.ID != nil {
			id = *app.ID
		}

		r, err := g.MotorRuntime.CreateResource("github.installation",
			"id", id,
			"appId", toInt64(app.AppID),
			"appSlug", toString(app.AppSlug),
			"createdAt", githubTimestamp(app.CreatedAt),
			"updatedAt", githubTimestamp(app.UpdatedAt),
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (g *lumiGithubTeam) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return "github.team/" + strconv.FormatInt(id, 10), nil
}

func newLumiGithubRepository(runtime *lumi.Runtime, repo *github.Repository) (interface{}, error) {
	var id int64
	if repo.ID != nil {
		id = *repo.ID
	}

	owner, err := runtime.CreateResource("github.user",
		"id", repo.GetOwner().GetID(),
		"login", repo.GetOwner().GetLogin(),
	)
	if err != nil {
		return nil, err
	}

	return runtime.CreateResource("github.repository",
		"id", id,
		"name", toString(repo.Name),
		"fullName", toString(repo.FullName),
		"description", toString(repo.Description),
		"homepage", toString(repo.Homepage),
		"createdAt", githubTimestamp(repo.CreatedAt),
		"updatedAt", githubTimestamp(repo.UpdatedAt),
		"archived", toBool(repo.Archived),
		"disabled", toBool(repo.Disabled),
		"private", toBool(repo.Private),
		"visibility", toString(repo.Visibility),
		"allowAutoMerge", toBool(repo.AllowAutoMerge),
		"allowForking", toBool(repo.AllowForking),
		"allowMergeCommit", toBool(repo.AllowMergeCommit),
		"allowRebaseMerge", toBool(repo.AllowRebaseMerge),
		"allowSquashMerge", toBool(repo.AllowSquashMerge),
		"hasIssues", toBool(repo.HasIssues),
		"organizationName", "",
		"defaultBranchName", toString(repo.DefaultBranch),
		"owner", owner,
	)
}

func (g *lumiGithubTeam) GetRepositories() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	teamID, err := g.Id()
	if err != nil {
		return nil, err
	}

	org, err := g.Organization()
	if err != nil {
		return nil, err
	}

	orgID, err := org.Id()
	if err != nil {
		return nil, err
	}

	repos, _, err := gt.Client().Teams.ListTeamReposByID(context.Background(), orgID, teamID, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	res := []interface{}{}
	for i := range repos {
		repo := repos[i]

		r, err := newLumiGithubRepository(g.MotorRuntime, repo)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (g *lumiGithubTeam) GetMembers() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	teamID, err := g.Id()
	if err != nil {
		return nil, err
	}

	org, err := g.Organization()
	if err != nil {
		return nil, err
	}

	orgID, err := org.Id()
	if err != nil {
		return nil, err
	}

	members, _, err := gt.Client().Teams.ListTeamMembersByID(context.Background(), orgID, teamID, &github.TeamListTeamMembersOptions{})
	if err != nil {
		return nil, err
	}

	res := []interface{}{}
	for i := range members {
		member := members[i]

		r, err := g.MotorRuntime.CreateResource("github.user",
			"id", toInt64(member.ID),
			"login", toString(member.Login),
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (g *lumiGithubUser) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return "github.user/" + strconv.FormatInt(id, 10), nil
}

func (g *lumiGithubUser) init(args *lumi.Args) (*lumi.Args, GithubUser, error) {
	if len(*args) > 3 {
		return args, nil, nil
	}

	if (*args)["login"] == nil {
		return nil, nil, errors.New("login required to fetch github user")
	}
	userLogin := (*args)["login"].(string)

	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, nil, err
	}

	user, _, err := gt.Client().Users.Get(context.Background(), userLogin)
	if err != nil {
		return nil, nil, err
	}

	(*args)["id"] = (*args)["id"]
	(*args)["login"] = toString(user.Login)
	(*args)["name"] = toString(user.Name)
	(*args)["email"] = toString(user.Email)
	(*args)["bio"] = toString(user.Bio)
	createdAt := &time.Time{}
	if user.CreatedAt != nil {
		createdAt = &user.CreatedAt.Time
	}
	(*args)["createdAt"] = createdAt
	updatedAt := &time.Time{}
	if user.UpdatedAt != nil {
		updatedAt = &user.UpdatedAt.Time
	}
	(*args)["updatedAt"] = updatedAt
	suspendedAt := &time.Time{}
	if user.SuspendedAt != nil {
		suspendedAt = &user.SuspendedAt.Time
	}
	(*args)["suspendedAt"] = suspendedAt
	(*args)["company"] = toString(user.Company)
	return args, nil, nil
}

func (g *lumiGithubCollaborator) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func (g *lumiGithubInstallation) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func (g *lumiGithubBranchprotection) id() (string, error) {
	return g.Id()
}

func (g *lumiGithubBranch) id() (string, error) {
	branchName, err := g.Name()
	if err != nil {
		return "", err
	}
	repoName, err := g.RepoName()
	if err != nil {
		return "", err
	}
	return repoName + "/" + branchName, nil
}

func (g *lumiGithubCommit) id() (string, error) {
	// the url is unique, e.g. "https://api.github.com/repos/vjeffrey/victoria-website/git/commits/7730d2707fdb6422f335fddc944ab169d45f3aa5"
	return g.Url()
}

func (g *lumiGithubReview) id() (string, error) {
	return g.Url()
}

func (g *lumiGithubRelease) id() (string, error) {
	return g.Url()
}

func (g *lumiGithubRepository) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func (g *lumiGithubRepository) init(args *lumi.Args) (*lumi.Args, GithubRepository, error) {
	if len(*args) > 2 {
		return args, nil, nil
	}
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, nil, err
	}

	// userLogin := user.GetLogin()
	org, err := gt.Organization()
	if err != nil {
		return nil, nil, err
	}

	owner := org.GetLogin()
	reponame := ""
	if x, ok := (*args)["name"]; ok {
		reponame = x.(string)
	} else {
		repo, err := gt.Repository()
		if err != nil {
			return nil, nil, err
		}
		reponame = *repo.Name
	}
	// return nil, nil, errors.New("Wrong type for 'path' in github.repository initialization, it must be a string")

	if owner != "" && reponame != "" {
		repo, _, err := gt.Client().Repositories.Get(context.Background(), owner, reponame)
		if err != nil {
			return nil, nil, err
		}

		owner, err := g.MotorRuntime.CreateResource("github.user",
			"id", repo.GetOwner().GetID(),
			"login", repo.GetOwner().GetLogin(),
		)
		if err != nil {
			return nil, nil, err
		}

		(*args)["id"] = toInt64(repo.ID)
		(*args)["name"] = toString(repo.Name)
		(*args)["fullName"] = toString(repo.FullName)
		(*args)["description"] = toString(repo.Description)
		(*args)["homepage"] = toString(repo.Homepage)
		(*args)["createdAt"] = githubTimestamp(repo.CreatedAt)
		(*args)["updatedAt"] = githubTimestamp(repo.UpdatedAt)
		(*args)["archived"] = toBool(repo.Archived)
		(*args)["disabled"] = toBool(repo.Disabled)
		(*args)["private"] = toBool(repo.Private)
		(*args)["visibility"] = toString(repo.Visibility)
		(*args)["allowAutoMerge"] = toBool(repo.AllowAutoMerge)
		(*args)["allowForking"] = toBool(repo.AllowForking)
		(*args)["allowMergeCommit"] = toBool(repo.AllowMergeCommit)
		(*args)["allowRebaseMerge"] = toBool(repo.AllowRebaseMerge)
		(*args)["allowSquashMerge"] = toBool(repo.AllowSquashMerge)
		(*args)["hasIssues"] = toBool(repo.HasIssues)
		(*args)["organizationName"] = ""
		(*args)["defaultBranchName"] = toString(repo.DefaultBranch)
		(*args)["owner"] = owner
	}

	return args, nil, nil
}

func (g *lumiGithubRepository) GetOpenMergeRequests() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.Name()
	if err != nil {
		return nil, err
	}
	orgName, err := g.OrganizationName()
	if err != nil {
		return nil, err
	}
	ownerName, err := g.Owner()
	if err != nil {
		return nil, err
	}
	ownerLogin, err := ownerName.Login()
	if err != nil {
		return nil, err
	}
	pulls, _, err := gt.Client().PullRequests.List(context.TODO(), ownerLogin, repoName, &github.PullRequestListOptions{State: "open"})
	if err != nil {
		log.Error().Err(err).Msg("unable to pull merge requests list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}

	res := []interface{}{}
	for i := range pulls {
		pr := pulls[i]

		labels, err := jsonToDictSlice(pr.Labels)
		if err != nil {
			return nil, err
		}
		owner, err := g.MotorRuntime.CreateResource("github.user",
			"id", toInt64(pr.User.ID),
			"login", toString(pr.User.Login),
		)
		if err != nil {
			return nil, err
		}

		assigneesRes := []interface{}{}
		for i := range pr.Assignees {
			assignee, err := g.MotorRuntime.CreateResource("github.user",
				"id", toInt64(pr.Assignees[i].ID),
				"login", toString(pr.Assignees[i].Login),
			)
			if err != nil {
				return nil, err
			}
			assigneesRes = append(assigneesRes, assignee)
		}

		r, err := g.MotorRuntime.CreateResource("github.mergeRequest",
			"id", toInt64(pr.ID),
			"number", toInt(pr.Number),
			"state", toString(pr.State),
			"labels", labels,
			"createdAt", pr.CreatedAt,
			"title", toString(pr.Title),
			"owner", owner,
			"assignees", assigneesRes,
			"organizationName", orgName,
			"repoName", repoName,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (g *lumiGithubMergeRequest) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func (g *lumiGithubRepository) GetBranches() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.Name()
	if err != nil {
		return nil, err
	}
	orgName, err := g.OrganizationName()
	if err != nil {
		return nil, err
	}
	owner, err := g.Owner()
	if err != nil {
		return nil, err
	}

	ownerLogin, err := owner.Login()
	if err != nil {
		return nil, err
	}

	repoDefaultBranchName, err := g.DefaultBranchName()
	if err != nil {
		return nil, err
	}

	branches, _, err := gt.Client().Repositories.ListBranches(context.TODO(), ownerLogin, repoName, &github.BranchListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to pull branches list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range branches {
		branch := branches[i]
		rc := branch.Commit
		lumiCommit, err := newLumiGithubCommit(g.MotorRuntime, rc, ownerLogin, repoName)
		if err != nil {
			return nil, err
		}

		defaultBranch := false
		if repoDefaultBranchName == toString(branch.Name) {
			defaultBranch = true
		}

		lumiBranch, err := g.MotorRuntime.CreateResource("github.branch",
			"name", branch.GetName(),
			"protected", branch.GetProtected(),
			"headCommit", lumiCommit,
			"organizationName", orgName,
			"repoName", repoName,
			"owner", owner,
			"isDefault", defaultBranch,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiBranch)
	}
	return res, nil
}

type githubDismissalRestrictions struct {
	Users []string `json:"users"`
	Teams []string `json:"teams"`
}

type githubRequiredPullRequestReviews struct {
	DismissalRestrictions *githubDismissalRestrictions `json:"dismissalRestrictions"`
	// Specifies if approved reviews are dismissed automatically, when a new commit is pushed.
	DismissStaleReviews bool `json:"dismissStaleReviews"`
	// RequireCodeOwnerReviews specifies if an approved review is required in pull requests including files with a designated code owner.
	RequireCodeOwnerReviews bool `json:"requireCodeOwnerReviews"`
	// RequiredApprovingReviewCount specifies the number of approvals required before the pull request can be merged.
	// Valid values are 1-6.
	RequiredApprovingReviewCount int `json:"requiredApprovingReviewCount"`
}

func (g *lumiGithubBranch) GetProtectionRules() (interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.RepoName()
	if err != nil {
		return nil, err
	}
	branchName, err := g.Name()
	if err != nil {
		return nil, err
	}
	owner, err := g.Owner()
	if err != nil {
		return nil, err
	}
	ownerName, err := owner.Login()
	if err != nil {
		log.Debug().Err(err).Msg("note: branch protection can only be accessed by admin users")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}

	branchProtection, _, err := gt.Client().Repositories.GetBranchProtection(context.TODO(), ownerName, repoName, branchName)
	if err != nil {
		// NOTE it is possible that the branch does not have any protection rules, therefore we don't return an error
		// TODO: figure out if the client has the permission to fetch the protection rules
		return nil, nil
	}
	rsc, err := jsonToDict(branchProtection.RequiredStatusChecks)
	if err != nil {
		return nil, err
	}

	var ghDismissalRestrictions *githubDismissalRestrictions

	if branchProtection.RequiredPullRequestReviews.DismissalRestrictions != nil {
		ghDismissalRestrictions = &githubDismissalRestrictions{
			Users: []string{},
			Teams: []string{},
		}

		for i := range branchProtection.RequiredPullRequestReviews.DismissalRestrictions.Teams {
			ghDismissalRestrictions.Teams = append(ghDismissalRestrictions.Teams, branchProtection.RequiredPullRequestReviews.DismissalRestrictions.Teams[i].GetName())
		}
		for i := range branchProtection.RequiredPullRequestReviews.DismissalRestrictions.Users {
			ghDismissalRestrictions.Users = append(ghDismissalRestrictions.Users, branchProtection.RequiredPullRequestReviews.DismissalRestrictions.Users[i].GetLogin())
		}
	}

	// we use a separate struct to ensure that the output is proper camelCase
	rprr, err := jsonToDict(githubRequiredPullRequestReviews{
		DismissStaleReviews:          branchProtection.RequiredPullRequestReviews.DismissStaleReviews,
		RequireCodeOwnerReviews:      branchProtection.RequiredPullRequestReviews.RequireCodeOwnerReviews,
		RequiredApprovingReviewCount: branchProtection.RequiredPullRequestReviews.RequiredApprovingReviewCount,
		DismissalRestrictions:        ghDismissalRestrictions,
	})
	if err != nil {
		return nil, err
	}

	ea, err := jsonToDict(branchProtection.EnforceAdmins)
	if err != nil {
		return nil, err
	}
	r, err := jsonToDict(branchProtection.Restrictions)
	if err != nil {
		return nil, err
	}
	rlh, err := jsonToDict(branchProtection.RequireLinearHistory)
	if err != nil {
		return nil, err
	}
	afp, err := jsonToDict(branchProtection.AllowForcePushes)
	if err != nil {
		return nil, err
	}
	ad, err := jsonToDict(branchProtection.AllowDeletions)
	if err != nil {
		return nil, err
	}
	rcr, err := jsonToDict(branchProtection.RequiredConversationResolution)
	if err != nil {
		return nil, err
	}

	sc, _, err := gt.Client().Repositories.GetSignaturesProtectedBranch(context.TODO(), ownerName, repoName, branchName)
	if err != nil {
		log.Debug().Err(err).Msg("note: branch protection can only be accessed by admin users")
		return nil, err
	}

	lumiBranchProtection, err := g.MotorRuntime.CreateResource("github.branchprotection",
		"id", repoName+"/"+branchName,
		"requiredStatusChecks", rsc,
		"requiredPullRequestReviews", rprr,
		"enforceAdmins", ea,
		"restrictions", r,
		"requireLinearHistory", rlh,
		"allowForcePushes", afp,
		"allowDeletions", ad,
		"requiredConversationResolution", rcr,
		"requiredSignatures", toBool(sc.Enabled),
	)
	if err != nil {
		return nil, err
	}
	return lumiBranchProtection, nil
}

func (g *lumiGitGpgSignature) id() (string, error) {
	sha, err := g.Sha()
	if err != nil {
		return "", err
	}
	return "git.gpgSignature/" + sha, nil
}

func newLumiGitGpgSignature(runtime *lumi.Runtime, sha string, a *github.SignatureVerification) (interface{}, error) {
	return runtime.CreateResource("git.gpgSignature",
		"sha", sha,
		"reason", a.GetReason(),
		"verified", a.GetVerified(),
		"payload", a.GetPayload(),
		"signature", a.GetSignature(),
	)
}

func (g *lumiGitCommitAuthor) id() (string, error) {
	sha, err := g.Sha()
	if err != nil {
		return "", err
	}
	return "git.commitAuthor/" + sha, nil
}

func newLumiGitAuthor(runtime *lumi.Runtime, sha string, a *github.CommitAuthor) (interface{}, error) {
	date := a.GetDate()
	return runtime.CreateResource("git.commitAuthor",
		"sha", sha,
		"name", a.GetName(),
		"email", a.GetEmail(),
		"date", &date,
	)
}

func (g *lumiGitCommit) id() (string, error) {
	sha, err := g.Sha()
	if err != nil {
		return "", err
	}
	return "git.commit/" + sha, nil
}

func newLumiGitCommit(runtime *lumi.Runtime, sha string, c *github.Commit) (interface{}, error) {
	// we have to pass-in the sha because the sha is often not set c.GetSHA()
	author, err := newLumiGitAuthor(runtime, sha, c.GetAuthor())
	if err != nil {
		return nil, err
	}

	committer, err := newLumiGitAuthor(runtime, sha, c.GetCommitter())
	if err != nil {
		return nil, err
	}

	signatureVerification, err := newLumiGitGpgSignature(runtime, sha, c.GetVerification())
	if err != nil {
		return nil, err
	}

	return runtime.CreateResource("git.commit",
		"sha", sha,
		"message", c.GetMessage(),
		"author", author,
		"committer", committer,
		"signatureVerification", signatureVerification,
	)
}

func newLumiGithubCommit(runtime *lumi.Runtime, rc *github.RepositoryCommit, owner string, repo string) (interface{}, error) {
	var githubAuthor interface{}
	var err error

	// if the github author is nil, we have to load the commit again
	if rc.Author == nil {
		gt, err := githubtransport(runtime.Motor.Transport)
		if err != nil {
			return nil, err
		}
		rc, _, err = gt.Client().Repositories.GetCommit(context.TODO(), owner, repo, rc.GetSHA(), nil)
		if err != nil {
			return nil, err
		}
	}

	if rc.Author != nil {
		githubAuthor, err = runtime.CreateResource("github.user", "id", toInt64(rc.Author.ID), "login", toString(rc.Author.Login))
		if err != nil {
			return nil, err
		}
	}
	var githubCommitter interface{}
	if rc.Committer != nil {
		githubCommitter, err = runtime.CreateResource("github.user", "id", toInt64(rc.Committer.ID), "login", toString(rc.Author.Login))
		if err != nil {
			return nil, err
		}
	}

	sha := rc.GetSHA()

	stats, err := jsonToDict(rc.GetStats())
	if err != nil {
		return nil, err
	}

	lumiGitCommit, err := newLumiGitCommit(runtime, sha, rc.Commit)
	if err != nil {
		return nil, err
	}

	return runtime.CreateResource("github.commit",
		"url", rc.GetURL(),
		"sha", sha,
		"author", githubAuthor,
		"committer", githubCommitter,
		"owner", owner,
		"repository", repo,
		"commit", lumiGitCommit,
		"stats", stats,
	)
}

func (g *lumiGithubRepository) GetCommits() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.Name()
	if err != nil {
		return nil, err
	}
	orgName, err := g.OrganizationName()
	if err != nil {
		return nil, err
	}
	ownerName, err := g.Owner()
	if err != nil {
		return nil, err
	}
	ownerLogin, err := ownerName.Login()
	if err != nil {
		return nil, err
	}
	commits, _, err := gt.Client().Repositories.ListCommits(context.TODO(), ownerLogin, repoName, &github.CommitsListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get commits list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range commits {
		rc := commits[i]
		lumiCommit, err := newLumiGithubCommit(g.MotorRuntime, rc, orgName, repoName)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiCommit)
	}
	return res, nil
}

func (g *lumiGithubMergeRequest) GetReviews() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.RepoName()
	if err != nil {
		return nil, err
	}
	orgName, err := g.OrganizationName()
	if err != nil {
		return nil, err
	}
	prID, err := g.Number()
	if err != nil {
		return nil, err
	}
	reviews, _, err := gt.Client().PullRequests.ListReviews(context.TODO(), orgName, repoName, int(prID), &github.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get reviews list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}

	res := []interface{}{}
	for i := range reviews {
		r := reviews[i]
		var user interface{}
		if r.User != nil {
			user, err = g.MotorRuntime.CreateResource("github.user", "id", toInt64(r.User.ID), "login", toString(r.User.Login))
			if err != nil {
				return nil, err
			}
		}
		lumiReview, err := g.MotorRuntime.CreateResource("github.review",
			"url", toString(r.HTMLURL),
			"state", toString(r.State),
			"authorAssociation", toString(r.AuthorAssociation),
			"user", user,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiReview)
	}

	return res, nil
}

func (g *lumiGithubMergeRequest) GetCommits() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.RepoName()
	if err != nil {
		return nil, err
	}
	orgName, err := g.OrganizationName()
	if err != nil {
		return nil, err
	}
	prID, err := g.Number()
	if err != nil {
		return nil, err
	}
	commits, _, err := gt.Client().PullRequests.ListCommits(context.TODO(), orgName, repoName, int(prID), &github.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get commits list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range commits {
		rc := commits[i]

		lumiCommit, err := newLumiGithubCommit(g.MotorRuntime, rc, orgName, repoName)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiCommit)
	}
	return res, nil
}

func (g *lumiGithubRepository) GetContributors() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.Name()
	if err != nil {
		return nil, err
	}
	ownerName, err := g.Owner()
	if err != nil {
		return nil, err
	}
	ownerLogin, err := ownerName.Login()
	if err != nil {
		return nil, err
	}
	contributors, _, err := gt.Client().Repositories.ListContributors(context.TODO(), ownerLogin, repoName, &github.ListContributorsOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get contributors list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range contributors {
		lumiUser, err := g.MotorRuntime.CreateResource("github.user",
			"id", toInt64(contributors[i].ID),
			"login", toString(contributors[i].Login),
		)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiUser)
	}
	return res, nil
}

func (g *lumiGithubRepository) GetCollaborators() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.Name()
	if err != nil {
		return nil, err
	}
	ownerName, err := g.Owner()
	if err != nil {
		return nil, err
	}
	ownerLogin, err := ownerName.Login()
	if err != nil {
		return nil, err
	}
	contributors, _, err := gt.Client().Repositories.ListCollaborators(context.TODO(), ownerLogin, repoName, &github.ListCollaboratorsOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get collaborator list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range contributors {

		lumiUser, err := g.MotorRuntime.CreateResource("github.user",
			"id", toInt64(contributors[i].ID),
			"login", toString(contributors[i].Login),
		)
		if err != nil {
			return nil, err
		}

		permissions := []string{}
		for k := range contributors[i].Permissions {
			permissions = append(permissions, k)
		}

		lumiContributor, err := g.MotorRuntime.CreateResource("github.collaborator",
			"id", toInt64(contributors[i].ID),
			"user", lumiUser,
			"permissions", sliceInterface(permissions),
		)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiContributor)
	}
	return res, nil
}

func (g *lumiGithubRepository) GetReleases() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.Name()
	if err != nil {
		return nil, err
	}
	ownerName, err := g.Owner()
	if err != nil {
		return nil, err
	}
	ownerLogin, err := ownerName.Login()
	if err != nil {
		return nil, err
	}
	releases, _, err := gt.Client().Repositories.ListReleases(context.TODO(), ownerLogin, repoName, &github.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get releases list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range releases {
		r := releases[i]
		lumiUser, err := g.MotorRuntime.CreateResource("github.release",
			"url", toString(r.HTMLURL),
			"name", toString(r.Name),
			"tagName", toString(r.TagName),
			"preRelease", toBool(r.Prerelease),
		)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiUser)
	}

	return res, nil
}

func (g *lumiGithubWebhook) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return "github.webhook/" + strconv.FormatInt(id, 10), nil
}

func (g *lumiGithubRepository) GetWebhooks() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	repoName, err := g.Name()
	if err != nil {
		return nil, err
	}
	owner, err := g.Owner()
	if err != nil {
		return nil, err
	}
	ownerLogin, err := owner.Login()
	if err != nil {
		return nil, err
	}

	hooks, _, err := gt.Client().Repositories.ListHooks(context.TODO(), ownerLogin, repoName, &github.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get hooks list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range hooks {
		h := hooks[i]
		config, err := jsonToDict(h.Config)
		if err != nil {
			return nil, err
		}

		lumiWebhook, err := g.MotorRuntime.CreateResource("github.webhook",
			"id", toInt64(h.ID),
			"name", toString(h.Name),
			"events", sliceInterface(h.Events),
			"config", config,
			"url", toString(h.URL),
			"name", toString(h.Name),
			"active", toBool(h.Active),
		)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiWebhook)
	}

	return res, nil
}

func (g *lumiGithubWorkflow) id() (string, error) {
	id, err := g.Id()
	if err != nil {
		return "", err
	}
	return "github.workflow/" + strconv.FormatInt(id, 10), nil
}

func (g *lumiGithubWorkflow) GetConfiguration() (interface{}, error) {
	// TODO: to leverage the runtime, get the file resource, how to define the dependency
	file, err := g.File()
	if err != nil {
		return nil, err
	}
	content, err := file.Content()
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(content), &data)
	if err != nil {
		return nil, err
	}

	return jsonToDict(data)
}

func (g *lumiGithubWorkflow) GetFile() (interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	filePath, err := g.Path()
	if err != nil {
		return nil, err
	}

	entry, ok := g.LumiResource().Cache.Load("_repositoryFullName")
	if !ok {
		return nil, errors.New("unable to get repository name")
	}

	fullName := entry.Data.(string)
	fullNameSplit := strings.Split(fullName, "/")
	ownerLogin := fullNameSplit[0]
	repoName := fullNameSplit[1]

	// TODO: no branch support yet
	// if we workflow is running for a branch only, we do not see from the response the branch name
	fileContent, _, _, err := gt.Client().Repositories.GetContents(context.Background(), ownerLogin, repoName, filePath, &github.RepositoryContentGetOptions{})
	if err != nil {
		// TODO: should this be an error
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	return newLumiGithubFile(g.MotorRuntime, ownerLogin, repoName, fileContent)
}

func (g *lumiGithubRepository) GetWorkflows() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}

	repoName, err := g.Name()
	if err != nil {
		return nil, err
	}
	owner, err := g.Owner()
	if err != nil {
		return nil, err
	}
	ownerLogin, err := owner.Login()
	if err != nil {
		return nil, err
	}

	fullName, err := g.FullName()
	if err != nil {
		return nil, err
	}

	workflows, _, err := gt.Client().Actions.ListWorkflows(context.Background(), ownerLogin, repoName, &github.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get hooks list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}

	res := []interface{}{}
	for i := range workflows.Workflows {
		w := workflows.Workflows[i]

		lumiWebhook, err := g.MotorRuntime.CreateResource("github.workflow",
			"id", toInt64(w.ID),
			"name", toString(w.Name),
			"path", toString(w.Path),
			"state", toString(w.State),
			"createdAt", githubTimestamp(w.CreatedAt),
			"updatedAt", githubTimestamp(w.UpdatedAt),
		)
		if err != nil {
			return nil, err
		}
		gw := lumiWebhook.(GithubWorkflow)
		gw.LumiResource().Cache.Store("_repositoryFullName", &lumi.CacheEntry{
			Data: fullName,
		})
		res = append(res, gw)
	}
	return res, nil
}

func newLumiGithubFile(runtime *lumi.Runtime, ownerName string, repoName string, content *github.RepositoryContent) (interface{}, error) {
	isBinary := false
	if toString(content.Type) == "file" {
		file := strings.Split(toString(content.Path), ".")
		if len(file) == 2 {
			isBinary = binaryFileTypes[file[1]]
		}
	}
	return runtime.CreateResource("github.file",
		"path", toString(content.Path),
		"name", toString(content.Name),
		"type", toString(content.Type),
		"sha", toString(content.SHA),
		"isBinary", isBinary,
		"ownerName", ownerName,
		"repoName", repoName,
	)
}

func (g *lumiGithubRepository) GetFiles() ([]interface{}, error) {
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.Name()
	if err != nil {
		return nil, err
	}

	owner, err := g.Owner()
	if err != nil {
		return nil, err
	}
	ownerLogin, err := owner.Login()
	if err != nil {
		return nil, err
	}
	_, dirContent, _, err := gt.Client().Repositories.GetContents(context.TODO(), ownerLogin, repoName, "", &github.RepositoryContentGetOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get contents list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range dirContent {
		lumiFile, err := newLumiGithubFile(g.MotorRuntime, ownerLogin, repoName, dirContent[i])
		if err != nil {
			return nil, err
		}
		res = append(res, lumiFile)

	}
	return res, nil
}

var binaryFileTypes = map[string]bool{
	"crx":    true,
	"deb":    true,
	"dex":    true,
	"dey":    true,
	"elf":    true,
	"o":      true,
	"so":     true,
	"iso":    true,
	"class":  true,
	"jar":    true,
	"bundle": true,
	"dylib":  true,
	"lib":    true,
	"msi":    true,
	"dll":    true,
	"drv":    true,
	"efi":    true,
	"exe":    true,
	"ocx":    true,
	"pyc":    true,
	"pyo":    true,
	"par":    true,
	"rpm":    true,
	"whl":    true,
}

func (g *lumiGithubFile) id() (string, error) {
	r, err := g.RepoName()
	if err != nil {
		return "", err
	}
	p, err := g.Path()
	if err != nil {
		return "", err
	}
	s, err := g.Sha()
	if err != nil {
		return "", err
	}
	return r + "/" + p + "/" + s, nil
}

func (g *lumiGithubFile) GetFiles() ([]interface{}, error) {
	fileType, err := g.Type()
	if err != nil {
		return nil, err
	}
	if fileType != "dir" {
		return nil, nil
	}
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return nil, err
	}
	repoName, err := g.RepoName()
	if err != nil {
		return nil, err
	}
	ownerName, err := g.OwnerName()
	if err != nil {
		return nil, err
	}
	path, err := g.Path()
	if err != nil {
		return nil, err
	}
	_, dirContent, _, err := gt.Client().Repositories.GetContents(context.TODO(), ownerName, repoName, path, &github.RepositoryContentGetOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get contents list")
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}
	res := []interface{}{}
	for i := range dirContent {
		isBinary := false
		if toString(dirContent[i].Type) == "file" {
			file := strings.Split(toString(dirContent[i].Path), ".")
			if len(file) == 2 {
				isBinary = binaryFileTypes[file[1]]
			}
		}
		lumiFile, err := g.MotorRuntime.CreateResource("github.file",
			"path", toString(dirContent[i].Path),
			"name", toString(dirContent[i].Name),
			"type", toString(dirContent[i].Type),
			"sha", toString(dirContent[i].SHA),
			"isBinary", isBinary,
			"ownerName", ownerName,
			"repoName", repoName,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, lumiFile)

	}
	return res, nil
}

func (g *lumiGithubFile) GetContent() (string, error) {
	fileType, err := g.Type()
	if err != nil {
		return "", err
	}
	if fileType == "dir" {
		return "", nil
	}
	gt, err := githubtransport(g.MotorRuntime.Motor.Transport)
	if err != nil {
		return "", err
	}
	repoName, err := g.RepoName()
	if err != nil {
		return "", err
	}
	ownerName, err := g.OwnerName()
	if err != nil {
		return "", err
	}
	path, err := g.Path()
	if err != nil {
		return "", err
	}
	fileContent, _, _, err := gt.Client().Repositories.GetContents(context.TODO(), ownerName, repoName, path, &github.RepositoryContentGetOptions{})
	if err != nil {
		log.Error().Err(err).Msg("unable to get contents list")
		if strings.Contains(err.Error(), "404") {
			return "", nil
		}
		return "", err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		if strings.Contains(err.Error(), "unsupported content encoding: none") {
			// TODO: i'm unclear why this error happens. the function checks for bas64 encoding and empty string encoding. if it's neither, it returns this error.
			// the error blocks the rest of the output, so we log it instead
			log.Error().Msgf("unable to get content for path %v", path)
			return "", nil
		}
	}
	return content, nil
}
