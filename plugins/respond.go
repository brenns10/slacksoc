package plugins

import "math/rand"
import "github.com/brenns10/slacksoc/lib"
import "github.com/mitchellh/mapstructure"
import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"

/*
An entry to associate a trigger with multiple potential replies.
*/
type response struct {
	Trigger string
	Replies []string
}

/*
This is a fairly simple plugin that allows you to configure triggers and
responses, nearly identical to traditional slackbot.
*/
type respond struct {
	Responses []response
}

/*
Creates a new Respond plugin. Don't bother calling this yourself, or even
manually registering it. Instead, use Register function for the core plugin lib.
*/
func newRespond(bot *lib.Bot, name string, config lib.PluginConfig) lib.Plugin {
	var respond respond
	err := mapstructure.Decode(config, &respond)
	if err != nil {
		return nil
	}
	bot.OnMessage("", respond.Respond)
	return &respond
}

func (r *respond) Describe() string {
	return "responds to triggers with randomly selected messages"
}

func (r *respond) Help() string {
	return "The Respond plugin  will listen to all messages and, if a trigger" +
		" message is heard, it will randomly select a response and send that" +
		" back to the channel."
}

func (r *respond) Respond(bot *lib.Bot, event *slack.MessageEvent) error {
	for _, resp := range r.Responses {
		if event.Msg.Text != resp.Trigger {
			continue
		}
		replyIndex := rand.Intn(len(resp.Replies))
		bot.RTM.SendMessage(bot.RTM.NewOutgoingMessage(resp.Replies[replyIndex],
			event.Msg.Channel))
		bot.Log.WithFields(logrus.Fields{
			"trigger": resp.Trigger, "reply": resp.Replies[replyIndex],
		}).Info("Respond trigger activated.")
		return nil
	}
	return nil
}
