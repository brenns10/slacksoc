package plugins

import "github.com/brenns10/slacksoc/lib"
import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"

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
		bot.Log.WithFields(logrus.Fields{
			"username": user.Name, "id": user.ID, "email": user.Profile.Email,
		}).Info("user")
	}
	return nil
}

func (d *debug) Channels(bot *lib.Bot, event *slack.MessageEvent) error {
	if event.Msg.Text != "channels" {
		return nil
	}
	for _, channel := range bot.GetChannels() {
		bot.Log.WithFields(logrus.Fields{
			"name": channel.Name, "id": channel.ID,
		}).Info("channel")
	}
	return nil
}

func (d *debug) Metadata(bot *lib.Bot, event *slack.MessageEvent) error {
	if event.Msg.Text != "metadata" {
		return nil
	}
	bot.Log.WithFields(logrus.Fields{
		"id": bot.Team.ID, "name": bot.Team.Name, "domain": bot.Team.Domain,
	}).Info("team")
	bot.Log.WithFields(logrus.Fields{
		"id": bot.User.ID, "name": bot.User.Name,
	}).Info("me")
	return nil
}

func (d *debug) Describe() string {
	return "several commands for seeing the internal state of the bot"
}

func (d *debug) Help() string {
	return "The Debug plugin contains several plugins for debugging the bot.\n" +
		"  users - log a list of users\n" +
		"  channels - log a list of users\n" +
		"  metadata - log the team and user data"
}

/*
Create a new debug plugin.
*/
func newDebug(bot *lib.Bot, _ string, _ lib.PluginConfig) lib.Plugin {
	d := &debug{}
	bot.OnMessage("", d.Users)
	bot.OnMessage("", d.Channels)
	bot.OnMessage("", d.Metadata)
	return d
}
