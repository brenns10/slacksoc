package plugins

/*
This file implements a "hot potato" game plugin, in which users pass the potato
to each other.
*/

import (
	"bytes"
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
}

type potatoGame struct {
	History []potatoEntry
	Unique  int
}

type hotPotato struct {
	// Configurable fields
	Timeout            int64
	DiversityThreshold float64

	// Private (non-configuration) fields
	name       string
	lock       sync.Mutex
	passRegexp *regexp.Regexp
	game       potatoGame
	timersSet  bool
	timer      *time.Timer
}

func newHotPotato(bot *lib.Bot, name string, cfg lib.PluginConfig) lib.Plugin {
	p := hotPotato{}
	p.name = name
	p.timersSet = false

	bot.Configure(cfg, &p, []string{"Timeout", "DiversityThreshold"})
	bot.GetState(name, &p.game) // in case a game already existed
	p.passRegexp = regexp.MustCompile(`(?i)pass the (?:hot )?potato to <@(U\w+)(\|\w+)?>`)

	bot.OnAddressedMatchExpr(p.passRegexp, p.locked(p.Pass))
	bot.OnAddressedMatch(`(?i)^give me the potato[!.]?$`, p.locked(p.Give))
	bot.OnAddressedMatch(`(?i)^who has the (?:hot )?potato[?.!]?$`,
		p.locked(p.Who))
	bot.OnAddressedMatch(`(?i)^potato history$`,
		p.locked(p.Had))
	bot.OnEvent("hello", p.Hello)

	return &p
}

/*
This function runs once we've connected to slack (for the first time) and there
was a pre-existing game on startup. It lets us restore timers and send some
apologies in case people tried to pass the potato while we were down.
*/
func (p *hotPotato) Hello(bot *lib.Bot, evt slack.RTMEvent) error {
	p.lock.Lock()

	if len(p.game.History) > 0 && !p.timersSet {
		// The bot started up and there was a game running!
		bot.Log.Info("HotPotato: preexisting game, resetting...")
		lastIdx := len(p.game.History) - 1
		entry := &p.game.History[lastIdx]
		duration := time.Duration(p.Timeout) * time.Minute
		endTime := entry.Received.Add(duration)
		remainingTime := endTime.Sub(time.Now())
		if remainingTime <= time.Duration(0) {
			bot.DirectMessage(entry.Uid, "Sorry, it looks like I crashed in "+
				"the middle of your game. The game has ended, but you can "+
				"start a new one if you'd like.",
			)
		}
		p.timer = time.AfterFunc(
			remainingTime, p.GameOver(bot, entry.Uid))
	}
	p.timersSet = true

	p.lock.Unlock()
	return nil
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
		"*slacksoc give me the potato* - starts a game if there's not one " +
		"happening\n" +
		"*slacksoc who has the potato* - tells you who has the potato, and " +
		"how long they have left until they need to pass it\n" +
		"*slacksoc potato history* - tells you who has had the potato over " +
		"the course of the current game\n" +
		"usage (in DMs):\n" +
		"*pass the potato to* _@username_ - passes to _@username_ if you " +
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
	for _, entry := range p.game.History {
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
	currentLen := len(p.game.History)
	return func() {
		p.lock.Lock()
		if len(p.game.History) != currentLen {
			// Someone else got the potato, but somehow the timer wasn't
			// canceled in time. Let's do the right thing: unlock and not send
			// any messages.
			p.lock.Unlock()
			return
		}
		bot.DirectMessage(uid, "Uh oh, you ran out of time. Game Over!")
		message := fmt.Sprintf("The game of hot potato ended with %s after "+
			"%d passes.", bot.Mention(bot.GetUserByID(uid)), currentLen)
		bot.Send(bot.GetChannelByName("random"), message)
		p.game.History = nil
		p.game.Unique = 0
		bot.UpdateState(p.name, &p.game)
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
	if len(p.game.History) == 0 {
		bot.Reply(evt, "There's no game happening right now! You could grab "+
			"the potato if you want.")
		return nil
	}
	lastIdx := len(p.game.History) - 1
	if evt.User != p.game.History[lastIdx].Uid {
		bot.Reply(evt, "You don't have the potato right now!")
		return nil
	}
	target := p.passRegexp.FindStringSubmatch(evt.Text)[1]
	if target == evt.User || target == "USLACKBOT" || target == bot.User.ID {
		bot.Reply(evt, "You can't pass the potato to them.")
		return nil
	}
	if p.userInHistory(target) {
		if float64(len(p.game.History)+1)/float64(p.game.Unique) > p.DiversityThreshold {
			bot.Reply(evt, "Try sending to someone new!")
			return nil
		}
	} else {
		p.game.Unique += 1
	}

	// add a history entry for this pass
	p.game.History[lastIdx].Passed = time.Now()
	newEntry := potatoEntry{
		Uid:      target,
		Received: time.Now(),
	}
	p.game.History = append(p.game.History, newEntry)
	bot.UpdateState(p.name, &p.game)

	// stop the old timer and start a new one
	p.timer.Stop()
	p.timer = time.AfterFunc(time.Duration(p.Timeout)*time.Minute,
		p.GameOver(bot, target))

	// notify the new person that they have the potato
	bot.DirectMessage(target, fmt.Sprintf(
		"%s passed you the hot potato :sweet_potato:! You "+
			"can pass it by replying to this DM: 'pass the potato "+
			"to @username'",
		bot.Mention(bot.GetUserByID(evt.User)),
	))
	// notify the sender that they have sent the potato
	bot.DirectMessage(evt.User, fmt.Sprintf(
		"Passed the potato to %s :sweet_potato:",
		bot.Mention(bot.GetUserByID(target)),
	))
	// and send a message to #random about it
	bot.Send(bot.GetChannelByName("random"), fmt.Sprintf(
		"%s passed the potato to %s :sweet_potato:",
		bot.Mention(bot.GetUserByID(evt.User)),
		bot.Mention(bot.GetUserByID(target))))

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
	if len(p.game.History) > 0 {
		bot.Reply(evt, "There is a game running right now.")
		return nil
	}

	// Add new entry to history
	newEntry := potatoEntry{
		Uid:      evt.User,
		Received: time.Now(),
	}
	p.game.History = append(p.game.History, newEntry)
	p.game.Unique = 1
	bot.UpdateState(p.name, &p.game)

	// Set a new timer.
	p.timer = time.AfterFunc(time.Duration(p.Timeout)*time.Minute,
		p.GameOver(bot, evt.User))

	// And send messages to people.
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
	if len(p.game.History) == 0 {
		bot.Reply(evt, "There's no game happening right now.")
		return nil
	}

	lastIdx := len(p.game.History) - 1
	lastEntry := p.game.History[lastIdx]
	user := bot.GetUserByID(lastEntry.Uid)
	deadline := lastEntry.Received.Add(time.Duration(p.Timeout) * time.Minute)
	bot.Reply(evt, fmt.Sprintf(
		"%s got the hot potato at %s. They have until %s to pass it. "+
			"The potato has been passed %d times.",
		bot.Mention(user), lastEntry.Received.Format("3:04 PM"),
		deadline.Format("3:04 PM"), len(p.game.History),
	))

	return nil
}

/*
Handles "potato history". Assumes we hold the lock.
*/
func (p *hotPotato) Had(bot *lib.Bot, evt *slack.MessageEvent) error {
	if len(p.game.History) == 0 {
		bot.Reply(evt, "There's no game happening right now.")
		return nil
	}
	var buffer bytes.Buffer
	buffer.WriteString(bot.User.Name)
	for _, entry := range p.game.History {
		buffer.WriteString(" - ")
		buffer.WriteString(bot.GetUserByID(entry.Uid).Name)
	}
	bot.Reply(evt, buffer.String())
	return nil
}
