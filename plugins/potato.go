package plugins

/*
This file implements a "hot potato" game plugin, in which users pass the potato
to each other.
*/

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/brenns10/slacksoc/lib"
	"github.com/nlopes/slack"
)

type potatoEntry struct {
	Uid      string
	Received time.Time
	Passed   time.Time
	Timer    *time.Timer
}

type hotPotato struct {
	// Configurable fields
	Timeout            int64
	DiversityThreshold float64

	// Private (non-configuration) fields
	lock       sync.Mutex
	history    []potatoEntry
	passRegexp *regexp.Regexp
	unique     int
}

func newHotPotato(bot *lib.Bot, _ string, cfg lib.PluginConfig) lib.Plugin {
	p := hotPotato{}

	bot.Configure(cfg, &p, []string{"Timeout", "DiversityThreshold"})
	p.passRegexp = regexp.MustCompile(`(?i)pass the (?:hot )?potato to <@(U\w+)(\|\w+)?>`)

	bot.OnAddressedMatchExpr(p.passRegexp, p.locked(p.Pass))
	bot.OnAddressedMatch(`(?i)^give me the potato[!.]?$`, p.locked(p.Give))
	bot.OnAddressedMatch(`(?i)^who has the (?:hot )?potato[?.!]?`,
		p.locked(p.Who))

	return &p
}

func (p *hotPotato) Describe() string {
	return "a game where you pass the hot potato"
}

func (p *hotPotato) Help() string {
	return "This is a game all about passing the potato. If you get the " +
		"potato, I'll PM you. Then you have to pass the potato to someone " +
		"before your timer runs out. If your timer runs out, you lose, and " +
		"you'll be publicly shamed in #random.\n" +
		"usage (in public channels):\n" +
		"**slacksoc give me the potato** - starts a game if there's not one " +
		"happening\n" +
		"**slacksoc who has the potato** - tells you who has the potato, and " +
		"how long they have left until they need to pass it\n" +
		"usage (in DMs):\n" +
		"**pass the potato to** _@username_ - passes to _@username_ if you " +
		"have the potato. They'll know who passed it to them"
}

/*
Takes a MessageHandler and wraps it with a locking statement so that all bot
events take the lock.
*/
func (p *hotPotato) locked(mh lib.MessageHandler) lib.MessageHandler {
	return func(bot *lib.Bot, evt *slack.MessageEvent) error {
		p.lock.Lock()
		rv := mh(bot, evt)
		p.lock.Unlock()
		return rv
	}
}

/*
Returns true if the user is already in the history.
*/
func (p *hotPotato) userInHistory(uid string) bool {
	for _, entry := range p.history {
		if entry.Uid == uid {
			return true
		}
	}
	return false
}

/*
Returns a callable function that will end the game for the given user. This
should be used with the Go Timer functionality. This function is capable of
detecting if the potato was already passed, and not ending the game in that
case.
*/
func (p *hotPotato) GameOver(bot *lib.Bot, uid string) func() {
	currentLen := len(p.history)
	return func() {
		p.lock.Lock()
		if len(p.history) != currentLen+1 {
			// the history entry is added after this is called
			return
		}
		bot.DirectMessage(uid, "Uh oh, you ran out of time. Game Over!")
		message := fmt.Sprintf("The game of hot potato ended with %s after "+
			"%d passes.", bot.Mention(bot.GetUserByID(uid)), currentLen+1)
		bot.Send(bot.GetChannelByName("random"), message)
		p.history = nil
		p.unique = 0
		p.lock.Unlock()
	}
}

/*
Handles the "pass the potato" command. Assumes that we hold the lock.
*/
func (p *hotPotato) Pass(bot *lib.Bot, evt *slack.MessageEvent) error {
	if !lib.IsDM(evt.Channel) {
		bot.React(evt, "no_entry_sign")
		return nil
	}
	if len(p.history) == 0 {
		bot.Reply(evt, "There's no game happening right now! You could grab "+
			"the potato if you want.")
		return nil
	}
	lastIdx := len(p.history) - 1
	if evt.User != p.history[lastIdx].Uid {
		bot.Reply(evt, "You don't have the potato right now!")
		return nil
	}
	target := p.passRegexp.FindStringSubmatch(evt.Text)[1]
	if target == evt.User || target == "USLACKBOT" || target == bot.User.ID {
		bot.Reply(evt, "You can't pass the potato to them.")
		return nil
	}
	if p.userInHistory(target) {
		if float64(len(p.history)+1)/float64(p.unique) > p.DiversityThreshold {
			bot.Reply(evt, "Try sending to someone new!")
			return nil
		}
	} else {
		p.unique += 1
	}

	// don't end the game with the last person!
	p.history[lastIdx].Timer.Stop()

	// add a history entry for this pass
	p.history[lastIdx].Passed = time.Now()
	newEntry := potatoEntry{
		Uid:      target,
		Received: time.Now(),
		Timer: time.AfterFunc(time.Duration(p.Timeout)*time.Minute,
			p.GameOver(bot, target)),
	}
	p.history = append(p.history, newEntry)

	// notify the new person that they have the potato
	bot.DirectMessage(target, fmt.Sprintf(
		"%s passed you the hot potato :sweet_potato:! You "+
			"can pass it by saying 'pass the potato to @username'",
		bot.Mention(bot.GetUserByID(evt.User)),
	))
	// notify the sender that they have sent the potato
	bot.DirectMessage(evt.User, fmt.Sprintf(
		"Passed the potato to %s :sweet_potato:",
		bot.Mention(bot.GetUserByID(target)),
	))

	return nil
}

/*
Handles the "give me the potato" command. Assumes that we hold the lock.
*/
func (p *hotPotato) Give(bot *lib.Bot, evt *slack.MessageEvent) error {
	if !lib.IsChannel(evt.Channel) {
		bot.Reply(evt, "why don't you ask me in a public channel?")
		return nil
	}
	if len(p.history) > 0 {
		bot.Reply(evt, "There is a game running right now.")
		return nil
	}
	newEntry := potatoEntry{
		Uid:      evt.User,
		Received: time.Now(),
		Timer: time.AfterFunc(time.Duration(p.Timeout)*time.Minute,
			p.GameOver(bot, evt.User)),
	}
	p.history = append(p.history, newEntry)
	p.unique = 1
	bot.Reply(evt,
		fmt.Sprintf("%s now has the hot potato :sweet_potato:. Let the game begin!",
			bot.Mention(bot.GetUserByID(evt.User))))
	bot.DirectMessage(evt.User, "You have the hot potato :sweet_potato:! "+
		"Say 'pass the potato to @username' to pass!")

	return nil
}

/*
Handles the "who has the potato" question. Assumes that we hold the lock.
*/
func (p *hotPotato) Who(bot *lib.Bot, evt *slack.MessageEvent) error {
	if len(p.history) == 0 {
		bot.Reply(evt, "There's no game happening right now.")
		return nil
	}

	lastIdx := len(p.history) - 1
	lastEntry := p.history[lastIdx]
	user := bot.GetUserByID(lastEntry.Uid)
	deadline := lastEntry.Received.Add(time.Duration(p.Timeout) * time.Minute)
	bot.Reply(evt, fmt.Sprintf(
		"%s got the hot potato at %s. They have until %s to pass it. "+
			"The potato has been passed %d times.",
		bot.Mention(user), lastEntry.Received.Format("3:05 PM"),
		deadline.Format("3:05 PM"), len(p.history),
	))

	return nil
}
