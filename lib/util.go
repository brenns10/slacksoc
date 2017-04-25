package lib

import "fmt"
import "regexp"
import "strings"

import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"

/*
A helper method which will reply to a message event with a message. This doesn't
block the main thread.
*/
func (bot *Bot) Reply(evt *slack.MessageEvent, msg string) {
	bot.RTM.SendMessage(bot.RTM.NewOutgoingMessage(msg, evt.Msg.Channel))
}

/*
A helper method for sending to any channel. You can do this with the underlying
slack library primitives, but this saves some typing and it could insulate
plugins from API changes.
*/
func (bot *Bot) Send(channelID string, msg string) {
	bot.RTM.SendMessage(bot.RTM.NewOutgoingMessage(msg, channelID))
}

/*
A helper method for sending direct messages. This does not block the main
thread, since it executes in a goroutine.
*/
func (bot *Bot) DirectMessage(uid string, msg string) {
	go func() {
		_, _, channel, err := bot.API.OpenIMChannel(uid)
		if err != nil {
			bot.Log.WithFields(logrus.Fields{
				"uid": uid,
				"msg": msg,
				"error": err,
			}).Error("Failed to send DM.")
		}
		bot.Send(channel, msg)
	}()
}

/*
A helper method which will react to a message event with a reaction. This
doesn't block the main thread.
*/
func (bot *Bot) React(evt *slack.MessageEvent, reaction string) {
	go func() {
		bot.API.AddReaction(reaction, slack.ItemRef{
			Channel: evt.Msg.Channel, Timestamp: evt.Msg.Timestamp,
		})
	}()
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

/*
The IsChannel function, similar to IsDM, IsFile, IsGroup, and IsUser, returns
true if the given Slack ID corresponds to a Channel (DM, File, Group, User,
respectively). This is as simple as checking the first character of the ID, but
these functions are more convenient and more readable.
*/
func IsChannel(id string) bool {
	return strings.HasPrefix(id, "C")
}

func IsDM(id string) bool {
	return strings.HasPrefix(id, "D")
}

func IsFile(id string) bool {
	return strings.HasPrefix(id, "F")
}

func IsGroup(id string) bool {
	return strings.HasPrefix(id, "G")
}

func IsUser(id string) bool {
	return strings.HasPrefix(id, "U")
}

/*
Given string s, parse a user mention and return the user ID associated with it.
This assumes that the user mention is the entire string. If there is no mention,
returns an empty string.
*/
func ParseUserMention(s string) string {
	expr := regexp.MustCompile(`<@(U\w+)(\|\w+)?>`)
	match := expr.FindStringSubmatchIndex(s)
	if match == nil {
		return ""
	}
	if match[0] != 0 || match[1] != len(s) {
		return ""
	}
	return s[match[2]:match[3]]
}
