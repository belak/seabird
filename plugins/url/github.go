package url

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/lrstanley/girc"
	"golang.org/x/oauth2"

	seabird "github.com/belak/go-seabird"
	"github.com/belak/go-seabird/plugins/utils"
)

func init() {
	seabird.RegisterPlugin("url/github", newGithubProvider)
}

type githubConfig struct {
	Token string
}

type githubProvider struct {
	api    *github.Client
	logger *logrus.Entry
}

var (
	githubUserRegex  = regexp.MustCompile(`^/([^/]+)$`)
	githubRepoRegex  = regexp.MustCompile(`^/([^/]+)/([^/]+)$`)
	githubIssueRegex = regexp.MustCompile(`^/([^/]+)/([^/]+)/issues/([^/]+)$`)
	githubPullRegex  = regexp.MustCompile(`^/([^/]+)/([^/]+)/pull/([^/]+)$`)
	githubGistRegex  = regexp.MustCompile(`^/([^/]+)/([^/]+)$`)

	githubPrefix = "[Github]"
)

func parseUserRepoNum(matches []string) (string, string, int, error) {
	if len(matches) != 4 {
		return "", "", 0, errors.New("Incorrect number of matches")
	}

	retInt, err := strconv.ParseInt(matches[3], 10, 32)
	if err != nil {
		return "", "", 0, err
	}

	return matches[1], matches[2], int(retInt), nil
}

func newGithubProvider(b *seabird.Bot, urlPlugin *Plugin) error {
	t := &githubProvider{
		logger: b.GetLogger(),
	}

	gc := &githubConfig{}
	err := b.Config("github", gc)
	if err != nil {
		return err
	}

	// Create an oauth2 client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gc.Token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	// Create a github client from the oauth2 client
	t.api = github.NewClient(tc)

	urlPlugin.RegisterProvider("github.com", t.githubCallback)
	urlPlugin.RegisterProvider("gist.github.com", t.gistCallback)

	return nil
}

func (t *githubProvider) githubCallback(c *girc.Client, e girc.Event, url *url.URL) bool {
	if githubUserRegex.MatchString(url.Path) {
		return t.getUser(c, e, url.Path)
	} else if githubRepoRegex.MatchString(url.Path) {
		return t.getRepo(c, e, url.Path)
	} else if githubIssueRegex.MatchString(url.Path) {
		return t.getIssue(c, e, url.Path)
	} else if githubPullRegex.MatchString(url.Path) {
		return t.getPull(c, e, url.Path)
	}

	return false
}

func (t *githubProvider) gistCallback(c *girc.Client, e girc.Event, url *url.URL) bool {
	if githubGistRegex.MatchString(url.Path) {
		return t.getGist(c, e, url.Path)
	}

	return false
}

// Jay Vana (@jsvana) at Facebook - Bio bio bio
var userTemplate = utils.TemplateMustCompile("githubUser", `
{{- if .user.Name -}}
{{ .user.Name }}
{{- with .user.Login }}(@{{ . }}){{ end -}}
{{- else if .user.Login -}}
@{{ .user.Login }}
{{- end -}}
{{- with .user.Company }} at {{ . }}{{ end -}}
{{- with .user.Bio }} - {{ . }}{{ end -}}
`)

func (t *githubProvider) getUser(c *girc.Client, e girc.Event, url string) bool {
	matches := githubUserRegex.FindStringSubmatch(url)
	if len(matches) != 2 {
		return false
	}

	user, _, err := t.api.Users.Get(context.TODO(), matches[1])
	if err != nil {
		t.logger.WithError(err).Error("Failed to get user from github")
		return false
	}

	return utils.RenderRespond(
		c, e, t.logger, userTemplate, githubPrefix,
		map[string]interface{}{
			"user": user,
		},
	)
}

// jsvana/alfred [PHP] (forked from belak/alfred) Last pushed to 2 Jan 2015 - Description, 1 fork, 2 open issues, 4 stars
var repoTemplate = utils.TemplateMustCompile("githubRepo", `
{{- .repo.FullName -}}
{{- with .repo.Language }} [{{ . }}]{{ end -}}
{{- if and .repo.Fork .repo.Parent }} (forked from {{ .repo.Parent.FullName }}){{ end }}
{{- with .repo.PushedAt }} Last pushed to {{ . | dateFormat "2 Jan 2006" }}{{ end }}
{{- with .repo.Description }} - {{ . }}{{ end }}
{{- with .repo.ForksCount }}, {{ prettifySuffix . }} {{ pluralizeWord . "fork" }}{{ end }}
{{- with .repo.OpenIssuesCount }}, {{ prettifySuffix . }} {{ pluralizeWord . "open issue" }}{{ end }}
{{- with .repo.StargazersCount }}, {{ prettifySuffix . }} {{ pluralizeWord . "star" }}{{ end }}
`)

func (t *githubProvider) getRepo(c *girc.Client, e girc.Event, url string) bool {
	matches := githubRepoRegex.FindStringSubmatch(url)
	if len(matches) != 3 {
		return false
	}

	user := matches[1]
	repoName := matches[2]
	repo, _, err := t.api.Repositories.Get(context.TODO(), user, repoName)

	if err != nil {
		t.logger.WithError(err).Error("Failed to get repo from github")
		return false
	}

	logger := t.logger.WithField("repo", repo)

	// If the repo doesn't have a name, we get outta there
	if repo.FullName == nil || *repo.FullName == "" {
		logger.Error("Invalid repo returned from github")
		return false
	}

	return utils.RenderRespond(
		c, e, logger, repoTemplate, githubPrefix,
		map[string]interface{}{
			"repo": repo,
		},
	)
}

// Issue #42 on belak/go-seabird [open] (assigned to jsvana) - Issue title [created 2 Jan 2015]
var issueTemplate = utils.TemplateMustCompile("githubIssue", `
Issue #{{ .issue.Number }} on {{ .user }}/{{ .repo }} [{{ .issue.State }}]
{{- with .issue.Assignee }} (assigned to {{ .Login }}){{ end }}
{{- with .issue.Title }} - {{ . }}{{ end }}
{{- with .issue.CreatedAt }} [created {{ . | dateFormat "2 Jan 2006" }}]{{ end }}
`)

func (t *githubProvider) getIssue(c *girc.Client, e girc.Event, url string) bool {
	matches := githubIssueRegex.FindStringSubmatch(url)
	user, repo, issueNum, err := parseUserRepoNum(matches)
	if err != nil {
		t.logger.WithError(err).Error("Failed to parse URL")
		return false
	}

	issue, _, err := t.api.Issues.Get(context.TODO(), user, repo, issueNum)
	if err != nil {
		t.logger.WithError(err).Error("Failed to get issue from github")
		return false
	}

	return utils.RenderRespond(
		c, e, t.logger, issueTemplate, githubPrefix,
		map[string]interface{}{
			"issue": issue,
			"user":  user,
			"repo":  repo,
		},
	)
}

// Pull request #59 on belak/go-seabird [open] - Title title title [created 4 Jan 2015], 1 commit, 4 comments, 2 changed files
var prTemplate = utils.TemplateMustCompile("githubPRTemplate", `
Pull request #{{ .pull.Number }} on {{ .user }}/{{ .repo }} [{{ .pull.State }}]
{{- with .pull.User.Login }} created by {{ . }}{{ end }}
{{- with .pull.Title }} - {{ . }}{{ end }}
{{- with .pull.CreatedAt }} [created {{ . | dateFormat "2 Jan 2006" }}]{{ end }}
{{- with .pull.Commits }}, {{ pluralize . "commit" }}{{ end }}
{{- with .pull.Comments }}, {{ pluralize . "comment" }}{{ end }}
{{- with .pull.ChangedFiles }}, {{ pluralize . "changed file" }}{{ end }}
`)

func (t *githubProvider) getPull(c *girc.Client, e girc.Event, url string) bool {
	matches := githubPullRegex.FindStringSubmatch(url)
	user, repo, pullNum, err := parseUserRepoNum(matches)
	if err != nil {
		t.logger.WithError(err).Error("Failed to parse URL")
		return false
	}

	pull, _, err := t.api.PullRequests.Get(context.TODO(), user, repo, int(pullNum))
	if err != nil {
		t.logger.WithError(err).Error("Failed to get github pr")
		return false
	}

	return utils.RenderRespond(
		c, e, t.logger, prTemplate, githubPrefix,
		map[string]interface{}{
			"user": user,
			"repo": repo,
			"pull": pull,
		},
	)
}

// Created 3 Jan 2015 by belak - Description description, 1 file, 3 comments
var gistTemplate = utils.TemplateMustCompile("gist", `
Created {{ .gist.CreatedAt | dateFormat "2 Jan 2006" }}
{{- with .gist.Owner.Login }} by {{ . }}{{ end }}
{{- with .gist.Description }} - {{ . }}{{ end }}
{{- with .gist.Comments }}, {{ pluralize . "comment" }}{{ end }}
`)

func (t *githubProvider) getGist(c *girc.Client, e girc.Event, url string) bool {
	matches := githubGistRegex.FindStringSubmatch(url)
	if len(matches) != 3 {
		return false
	}

	id := matches[2]
	gist, _, err := t.api.Gists.Get(context.TODO(), id)
	if err != nil {
		t.logger.WithError(err).Error("Failed to get gist")
		return false
	}

	return utils.RenderRespond(
		c, e, t.logger, gistTemplate, githubPrefix,
		map[string]interface{}{
			"gist": gist,
		},
	)
}
