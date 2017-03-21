package lib

import "regexp"

import "github.com/nlopes/slack"

/*
EventHandler is a function type which handles an single Slack event from the
RTM API. Although a complete description of the Slack API is out of scope, you
can find information about the RTM API and the different types of events which
are sent in their documentation: https://api.slack.com/rtm

The EventHandler receives a pointer to the bot, as well as a copy of the slack
library's RTMEvent. This structure contains a Data pointer, which may point to
one of many types of structs in the library (those ending in -Event). A typical
first step is to use a type assertion to retrieve the actual event data from
this RTMEvent object. For documentation on the Go slack library types, see:
https://godoc.org/github.com/nlopes/slack

Finally, an important thing to note is that EventHandlers, like any function in
Go, may be methods with receivers, or functions that have closures. So, a plugin
could register one of its own methods as an event handler, and thus it would
always be called with the correct pointer to its plugin data.

Register these with bot.OnEvent()
*/
type EventHandler func(bot *Bot, evt slack.RTMEvent) error

/*
MessageHandler is a specialization of EventHandler. This takes care of some
boilerplate code for the common case, that you are implementing a handler for
the Slack "message" event type: https://api.slack.com/events/message

Register these with bot.OnMessage(), bot.OnAddressed(), bot.OnMatch(), or
bot.OnMatchExpr(), depending on what you want.
*/
type MessageHandler func(bot *Bot, msg *slack.MessageEvent) error

/*
CommandHandler is a further specialization of MessageHandler. It receives a
list of arguments. These arguments have been parsed out of the message, and they
do not include the part of the message that is addressed to the bot.

Register these with bot.OnCommand().
*/
type CommandHandler func(bot *Bot, msg *slack.MessageEvent, args []string) error

/*
Return a message handler which unconditionally responds with the given message.
*/
func Reply(msg string) MessageHandler {
	return func(bot *Bot, evt *slack.MessageEvent) error {
		bot.Reply(evt, msg)
		return nil
	}
}

/*
Return a message handler which unconditionally reacts with the given reaction.
*/
func React(rxn string) MessageHandler {
	return func(bot *Bot, evt *slack.MessageEvent) error {
		bot.React(evt, rxn)
		return nil
	}
}

/*
Same as IfMatch, but with a compiled expression.
*/
func IfMatchExpr(re *regexp.Regexp, mh MessageHandler) MessageHandler {
	return func(bot *Bot, evt *slack.MessageEvent) error {
		match := re.FindStringIndex(evt.Msg.Text)
		if len(match) > 0 {
			return mh(bot, evt)
		} else {
			return nil
		}
	}
}

/*
Return a message handler which will call another handler if the handler matches
a Regexp.
*/
func IfMatch(re string, mh MessageHandler) MessageHandler {
	return IfMatchExpr(regexp.MustCompile(re), mh)
}
