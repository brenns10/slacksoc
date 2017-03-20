package plugins

import "math/rand"
import "regexp"

import "github.com/brenns10/slacksoc/lib"
import "github.com/mitchellh/mapstructure"
import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"

/*
An entry to associate a trigger with multiple potential replies.
*/
type response struct {
	Trigger string
	trigger *regexp.Regexp
	Replies []string
	Reacts  []string
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
	for i, resp := range respond.Responses {
		respond.Responses[i].trigger = regexp.MustCompile(resp.Trigger)
		respond.Responses[i].trigger.Longest() // leftmost longest match
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
		// must be a full match
		if resp.trigger.FindString(event.Msg.Text) != event.Msg.Text {
			continue
		}
		var reply, react string
		if len(resp.Replies) > 0 {
			replyIndex := rand.Intn(len(resp.Replies))
			reply = resp.Replies[replyIndex]
			bot.Reply(event, reply)
		}
		if len(resp.Reacts) > 0 {
			reactIndex := rand.Intn(len(resp.Reacts))
			react = resp.Reacts[reactIndex]
			bot.React(event, react)

		}
		bot.Log.WithFields(logrus.Fields{
			"trigger": resp.Trigger, "reply": reply, "react": react,
		}).Info("Respond trigger activated.")
		return nil
	}
	return nil
}
