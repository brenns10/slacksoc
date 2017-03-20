package plugins

/*
This file provides a plugin for building Github integrations into
github.com/brenns10/slacksoc.

It is based on Andrew Mason's GitHub plugin for his slack bot library:
github.com/ajm188/slack
*/

import (
	"context"
	"strings"
	"time"

	"github.com/brenns10/slacksoc/lib"
	"github.com/google/go-github/github"
	"github.com/mitchellh/mapstructure"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	ghAuth "golang.org/x/oauth2/github" // have to rename so we don't have 2 "github"s
)

/*
Contains the OAuth "stuff" necessary to connect, and a client to actually
perform operations.
*/
type ghPlugin struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	client       *github.Client
}

func newGitHub(bot *lib.Bot, _ string, cfg lib.PluginConfig) lib.Plugin {
	g := ghPlugin{}
	err := mapstructure.Decode(cfg, &g)
	if err != nil {
		bot.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Error creating GitHub client.")
	}
	g.client = g.createClient()
	bot.OnCommand("issue", g.Issue)
	return &g
}

/*
Create a GitHub client from the configured information in the plugin.
*/
func (p *ghPlugin) createClient() *github.Client {
	var noExpire time.Time // this sets noExpire to the zero Time value
	config := &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Endpoint:     ghAuth.Endpoint,
		RedirectURL:  "",
		Scopes:       []string{"repo"},
	}
	token := &oauth2.Token{
		AccessToken:  p.AccessToken,
		TokenType:    "", // uhhh
		RefreshToken: "",
		Expiry:       noExpire,
	}
	return github.NewClient(config.Client(oauth2.NoContext, token))
}

func (p *ghPlugin) Describe() string {
	return "create GitHub issues"
}

func (p *ghPlugin) Help() string {
	return "usage:\n" +
		"**issue me** _owner/repo title [body [assignee]]_ - create a " +
		"GitHub issue\n"
}

/*
This plugin asynchronously creates a GitHub issue. The command looks like this:

    slacksoc issue [me] owner/repo "title" ["body" [assignee]]
*/
func (p *ghPlugin) Issue(bot *lib.Bot, evt *slack.MessageEvent, args []string) error {
	// goroutine is asynchronous so that we don't block the main thread
	go func() {
		// standardize the args to start at 0
		if len(args) >= 2 && args[1] == "me" {
			args = args[2:]
		} else {
			args = args[1:]
		}

		// help text if not enough stuff
		if len(args) < 2 {
			bot.Reply(evt, p.Help())
			return
		}

		// get necessary arguments
		var title, assignee *string
		var body string
		ownerRepo := strings.Split(args[0], "/")
		if len(ownerRepo) != 2 {
			bot.Reply(evt, "error: first argument should be owner/repo")
			return
		}
		owner := ownerRepo[0]
		repo := ownerRepo[1]
		user := bot.GetUserByID(evt.Msg.User)
		name := user.Name
		if user.RealName != "" {
			name += " (" + user.RealName + ")"
		}
		title = &args[1]
		if len(args) >= 3 {
			body = args[2]
			body += "\n\nCreated via Slack on behalf of " + name
		} else {
			body = "Created via Slack on behalf of " + name
		}
		if len(args) >= 4 {
			assignee = &args[3]
		}
		issueState := "open"

		// and send the request
		request := github.IssueRequest{
			Title:     title,
			Body:      &body,
			Labels:    nil,
			Assignee:  assignee,
			State:     &issueState,
			Milestone: nil,
		}
		issue, _, err := p.client.Issues.Create(context.TODO(), owner, repo, &request)
		logEntry := bot.Log.WithFields(logrus.Fields{
			"title": *title, "body": body, "assignee": *assignee, "owner": owner,
			"repos": repo,
		})
		if err != nil {
			bot.Reply(evt, "Error creating the issue: "+err.Error())
			logEntry.Error("Error creating a GitHub issue.")
		} else {
			bot.Reply(evt, *issue.HTMLURL)
			logEntry.Info("Created a GitHub issue.")
		}
	}()
	return nil
}
