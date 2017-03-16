package lib

import "github.com/nlopes/slack"

/*
Represents a function that is called on an event.
*/
type EventHandler func(bot *Bot, evt slack.RTMEvent) error

/*
This is a happy simplification for message events.
*/
type MessageHandler func(bot *Bot, msg *slack.MessageEvent) error
