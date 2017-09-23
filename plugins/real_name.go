package plugins

import "fmt"

import "github.com/brenns10/slacksoc/lib"
import "github.com/nlopes/slack"

type realName struct {
	Channel string
}

func (r *realName) RealName(bot *lib.Bot, event *slack.MessageEvent) error {
	if bot.GetChannelByID(event.Channel) != r.Channel {
		return nil
	}
	user := bot.GetUserByID(event.User)
	if user.RealName != "" {
		return nil
	}
	text := "Please set your real name fields. " +
		"https://%s.slack.com/team/%s. " +
		"Then click \"Edit\"."
	text = fmt.Sprintf(text, bot.Team.Domain, user.Name)
	bot.DirectMessage(user.ID, text)
	return nil
}

func (r *realName) Describe() string {
	return "makes people set real name fields"
}

func (r *realName) Help() string {
	return "The RealName plugin sends DMs to people who join #general " +
		"without setting their real name field."
}

func newRealName(bot *lib.Bot, _ string, cfg lib.PluginConfig) lib.Plugin {
	r := &realName{}
	bot.Configure(cfg, &r, []string{"Channel"})
	bot.OnMessage("channel_join", r.RealName)
	return r
}
