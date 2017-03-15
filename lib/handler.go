package lib

import "github.com/nlopes/slack"

type HandlerCallback func(bot *Bot, evt slack.RTMEvent, data interface{})
type HandlerCondition func(bot *Bot, evt slack.RTMEvent) bool

type Handler interface {
	Handle(bot *Bot, evt slack.RTMEvent)
	GetPlugin() *Plugin
}

type HandlerFunc struct {
	Plugin  *Plugin
	Handler HandlerCallback
}

func (hf *HandlerFunc) Handle(bot *Bot, evt slack.RTMEvent) {
	hf.Handler(bot, evt, hf.Plugin)
}

func (hf *HandlerFunc) GetPlugin() *Plugin {
	return hf.Plugin
}

func NewHandlerFunc(plugin *Plugin, function HandlerCallback) *HandlerFunc {
	return &HandlerFunc{
		Plugin:  plugin,
		Handler: function,
	}
}

type ConditionHandler struct {
	Condition HandlerCondition
	Handler   Handler
}

func (ch *ConditionHandler) Handle(bot *Bot, evt slack.RTMEvent) {
	if ch.Condition(bot, evt) {
		ch.Handler.Handle(bot, evt)
	}
}

func (ch *ConditionHandler) GetPlugin() *Plugin {
	return ch.Handler.GetPlugin()
}

func NewConditionHandler(handler Handler, condition HandlerCondition) *ConditionHandler {
	return &ConditionHandler{
		Condition: condition,
		Handler:   handler,
	}
}
