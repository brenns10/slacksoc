package plugins

import "fmt"
import "strings"

import "github.com/brenns10/slacksoc/lib"
import "github.com/hacsoc/golove/love"
import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"

type lov struct {
	client love.Client
}

func usernameForUser(user *slack.User) string {
	email := user.Profile.Email
	if email == "" {
		return email
	}
	return email[:strings.Index(email, "@")]
}

func usernameForString(bot *lib.Bot, arg string) string {
	uid := lib.ParseUserMention(arg)
	if uid != "" {
		return usernameForUser(bot.GetUserByID(uid))
	} else {
		return arg
	}
}

func (l *lov) Love(bot *lib.Bot, evt *slack.MessageEvent, args []string) error {
	// the whole thing is done asynchronously due to the API call, so we do not
	// block the main slacksoc goroutine
	go func() {
		if len(args) <= 2 {
			bot.Reply(evt, l.Help())
			return
		}
		usernames := make([]string, 0, len(args)-2)
		for _, arg := range args[1 : len(args)-1] {
			username := usernameForString(bot, arg)
			// deal with possible error getting a username
			if username == "" {
				bot.Reply(evt, fmt.Sprintf("Sorry, we had trouble turning "+
					"\"%s\" into a Case ID.", arg))
				return
			}
			usernames = append(usernames, usernameForString(bot, arg))
		}
		// deal with possible error getting a username
		sender := usernameForUser(bot.GetUserByID(evt.Msg.User))
		if sender == "" {
			bot.Reply(evt, "Sorry, we couldn't determine your username. Do you"+
				" have your email set in your profile?")
			return
		}
		entry := bot.Log.WithFields(logrus.Fields{
			"usernames": usernames, "sender": sender,
			"message": args[len(args)-1],
		})
		err := l.client.SendLoves(sender, usernames, args[len(args)-1])
		if err != nil {
			entry.Error(err)
			if strings.HasPrefix(err.Error(), "Love API Error: ") {
				// API errors are safe and contain user info
				bot.Reply(evt, err.Error())
			} else {
				// HTTP/other errors may contain sensitive info
				bot.Reply(evt, "An error occurred sending love. Consult slacksoc's"+
					" logs for more details.")
			}
		} else {
			entry.Info("sent love")
			bot.Reply(evt, ":sparkling_heart:")
		}
	}()
	return nil
}

func (l *lov) Describe() string {
	return "a command for sending CWRU love"
}

func (l *lov) Help() string {
	return "Command syntax:\n\n" +
		"*love _user_ [_user_ ...]* \"message\"\n" +
		"_user_ may be an @mentioned slack username. In this case, we will " +
		"get their Case ID from the email address in their Slack profile.\n" +
		"_user_ may also be just a Case ID\n\n" +
		"Note that slacksoc commands are parsed according to similar rules as" +
		" bash shell commands."
}

func newLove(bot *lib.Bot, _ string, cfg lib.PluginConfig) lib.Plugin {
	d := &lov{}
	bot.Configure(cfg, &d.client, []string{"ApiKey", "BaseUrl"})
	bot.OnCommand("love", d.Love)
	return d
}
