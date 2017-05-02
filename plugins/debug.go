package plugins

import "fmt"
import "strconv"

import "github.com/brenns10/slacksoc/lib"
import "github.com/nlopes/slack"

type debugConfig struct {
	Trusted []string
}

type debugState struct {
	State int64
}

type debug struct {
	name   string
	Config debugConfig
	State  debugState
}

func (d *debug) trustedCommand(ch lib.CommandHandler) lib.CommandHandler {
	return func(bot *lib.Bot, event *slack.MessageEvent, args []string) error {
		if lib.Contains(d.Config.Trusted, bot.GetUserByID(event.User).Name) {
			return ch(bot, event, args)
		}
		bot.React(event, "no_entry_sign")
		return nil
	}
}

func (d *debug) trustedHandler(mh lib.MessageHandler) lib.MessageHandler {
	return func(bot *lib.Bot, event *slack.MessageEvent) error {
		if lib.Contains(d.Config.Trusted, bot.GetUserByID(event.User).Name) {
			return mh(bot, event)
		}
		bot.React(event, "no_entry_sign")
		return nil
	}
}

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

func (d *debug) StateCmd(bot *lib.Bot, evt *slack.MessageEvent, args []string) error {
	if len(args) <= 1 {
		bot.Reply(evt, fmt.Sprintf("state is %d", d.State.State))
	} else {
		n, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			bot.Reply(evt, "Couldn't parse that number.")
		} else {
			d.State.State = n
			bot.UpdateState(d.name, d.State)
			bot.Reply(evt, "State has been updated.")
		}
	}
	return nil
}

func (d *debug) PM(bot *lib.Bot, evt *slack.MessageEvent) error {
	bot.DirectMessage(evt.User, "hi there!")
	return nil
}

func (d *debug) Describe() string {
	return "several commands for seeing the internal state of the bot"
}

func (d *debug) Help() string {
	return "The Debug plugin contains several commands for debugging the bot.\n" +
		"*slacksoc users:* reply with list of users\n" +
		"*slacksoc channels:* log a list of users\n" +
		"*slacksoc metadata:* log the team and user data\n" +
		"*slacksoc info:* tells you what channel you're in, etc\n" +
		"*slacksoc id _target_*: return Slack ID of something\n" +
		"*slacksoc state _[number]_*: set or get a persisted state number\n" +
		"*slacksoc pm me*: request a PM\n" +
		"*slacksoc version*: tells the slacksoc version"
}

/*
Create a new debug plugin.
*/
func newDebug(bot *lib.Bot, name string, cfg lib.PluginConfig) lib.Plugin {
	d := &debug{}
	d.name = name
	bot.Configure(cfg, &d.Config, []string{"Trusted"})
	bot.GetState(name, &d.State)
	bot.OnAddressedMatch("^users$", d.trustedHandler(d.Users))
	bot.OnAddressedMatch("^channels$", d.trustedHandler(d.Channels))
	bot.OnAddressedMatch("^metadata$", d.trustedHandler(d.Metadata))
	bot.OnAddressedMatch("^info$", d.trustedHandler(d.Info))
	bot.OnAddressedMatch("^version$", lib.Reply("My version is 1.1.2"))
	bot.OnCommand("id", d.Id)
	bot.OnCommand("state", d.StateCmd)
	bot.OnAddressedMatch("^pm me$", d.PM)
	return d
}
