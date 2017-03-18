package lib

import "github.com/nlopes/slack"

/*
A helper method which will reply to a message event with a message.
*/
func (bot *Bot) Reply(evt *slack.MessageEvent, msg string) {
	bot.RTM.SendMessage(bot.RTM.NewOutgoingMessage(msg, evt.Msg.Channel))
}

/*
A helper method which will react to a message event with a reaction.
*/
func (bot *Bot) React(evt *slack.MessageEvent, reaction string) {
	bot.API.AddReaction(reaction, slack.ItemRef{
		Channel: evt.Msg.Channel, Timestamp: evt.Msg.Timestamp,
	})
}
