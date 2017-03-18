package plugins

import "fmt"

import "github.com/brenns10/slacksoc/lib"
import "github.com/nlopes/slack"

/*
This is just for the plugin, we don't actually have any state.
*/
type debug struct{}

func (d *debug) Users(bot *lib.Bot, event *slack.MessageEvent) error {
	if event.Msg.Text != "users" {
		return nil
	}
	users := bot.GetUsers()
	for _, user := range users {
		bot.Reply(event, fmt.Sprintf("user: id=%s, username=%s, email=%s",
			user.ID, user.Name, user.Profile.Email))
	}
	return nil
}

func (d *debug) Channels(bot *lib.Bot, event *slack.MessageEvent) error {
	if event.Msg.Text != "channels" {
		return nil
	}
	for _, channel := range bot.GetChannels() {
		bot.Reply(event, fmt.Sprintf("channel: id=%s, name=%s", channel.ID,
			channel.Name))
	}
	return nil
}

func (d *debug) Metadata(bot *lib.Bot, event *slack.MessageEvent) error {
	if event.Msg.Text != "metadata" {
		return nil
	}
	bot.Reply(event, fmt.Sprintf("team: id=%s, name=%s, domain=%s",
		bot.Team.ID, bot.Team.Name, bot.Team.Domain))
	bot.Reply(event, fmt.Sprintf("me: id=%s, name=%s", bot.User.ID,
		bot.User.Name))
	return nil
}

func (d *debug) Debug(bot *lib.Bot, event *slack.MessageEvent) error {
	if event.Msg.Text != "debug" {
		return nil
	}
	bot.React(event, "dope")
	return nil
}

func (d *debug) Info(bot *lib.Bot, event *slack.MessageEvent) error {
	if event.Msg.Text != "info" {
		return nil
	}
	bot.Reply(event, bot.MentionI(event.Msg.User)+": we are in "+
		bot.SayChannelI(event.Msg.Channel)+", tell "+
		bot.SpecialMention("everyone"))
	return nil
}

func (d *debug) Describe() string {
	return "several commands for seeing the internal state of the bot"
}

func (d *debug) Help() string {
	return "The Debug plugin contains several plugins for debugging the bot.\n" +
		"  users - log a list of users\n" +
		"  channels - log a list of users\n" +
		"  metadata - log the team and user data\n" +
		"  debug - reacts to the message with :dope:\n" +
		"  info - tells you what channel you're in, etc"
}

/*
Create a new debug plugin.
*/
func newDebug(bot *lib.Bot, _ string, _ lib.PluginConfig) lib.Plugin {
	d := &debug{}
	bot.OnMessage("", d.Users)
	bot.OnMessage("", d.Channels)
	bot.OnMessage("", d.Metadata)
	bot.OnMessage("", d.Debug)
	bot.OnMessage("", d.Info)
	return d
}
