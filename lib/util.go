package lib

import "fmt"

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

/*
Construct a string to @mention a user.
*/
func (bot *Bot) Mention(user *slack.User) string {
	return fmt.Sprintf("<@%s>", user.Name)
}

/*
Construct a string to @mention a user, given username.
*/
func (bot *Bot) MentionN(username string) string {
	return bot.Mention(bot.GetUserByName(username))
}

/*
Construct a string to @mention a user, given user ID.
*/
func (bot *Bot) MentionI(id string) string {
	return bot.Mention(bot.GetUserByID(id))
}

/*
Construct a string for a special mention - @channel, @here, @group, @everyone.
*/
func (bot *Bot) SpecialMention(target string) string {
	return fmt.Sprintf("<!%s>", target)
}

/*
Construct a string to say a #channel, given its id.
*/
func (bot *Bot) SayChannelI(id string) string {
	return fmt.Sprintf("<#%s>", id)
}

/*
Construct a string to say a #channel, given the channel name.
*/
func (bot *Bot) SayChannelN(name string) string {
	return bot.SayChannelI(bot.GetChannelByName(name))
}
