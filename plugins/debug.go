package plugins

import "fmt"

import "github.com/brenns10/slacksoc/lib"
import "github.com/nlopes/slack"

/*
This is just for the plugin, we don't actually have any state.
*/
type debug struct{}

func (d *debug) Users(bot *lib.Bot, event *slack.MessageEvent) error {
	users := bot.GetUsers()
	for _, user := range users {
		bot.Reply(event, fmt.Sprintf("user: id=%s, username=%s, email=%s",
			user.ID, user.Name, user.Profile.Email))
	}
	return nil
}

func (d *debug) Channels(bot *lib.Bot, event *slack.MessageEvent) error {
	for _, channel := range bot.GetChannels() {
		bot.Reply(event, fmt.Sprintf("channel: id=%s, name=%s", channel.ID,
			channel.Name))
	}
	return nil
}

func (d *debug) Metadata(bot *lib.Bot, event *slack.MessageEvent) error {
	bot.Reply(event, fmt.Sprintf("team: id=%s, name=%s, domain=%s",
		bot.Team.ID, bot.Team.Name, bot.Team.Domain))
	bot.Reply(event, fmt.Sprintf("me: id=%s, name=%s", bot.User.ID,
		bot.User.Name))
	return nil
}

func (d *debug) Info(bot *lib.Bot, event *slack.MessageEvent) error {
	bot.Reply(event, bot.MentionI(event.Msg.User)+": we are in "+
		bot.SayChannelI(event.Msg.Channel)+", tell "+
		bot.SpecialMention("everyone"))
	return nil
}

func (d *debug) Id(bot *lib.Bot, evt *slack.MessageEvent, args []string) error {
	if len(args) <= 1 {
		bot.Reply(evt, "What do you want me to id?")
		return nil
	}
	if args[1] == "me" {
		bot.Reply(evt, evt.Msg.User)
	} else {
		bot.Reply(evt, "Sorry, I only know how to 'id me'")
	}
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
	bot.OnMatch("users", d.Users)
	bot.OnMatch("channels", d.Channels)
	bot.OnMatch("metadata", d.Metadata)
	bot.OnMatch("debug", lib.React("dope"))
	bot.OnMatch("info", d.Info)
	bot.OnCommand("id", d.Id)
	return d
}
