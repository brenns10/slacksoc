package lib

import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"

/*
This type simply represents a Channel entry that can be returned by the Bot.
Unfortunately, due to the way the RTM API works, we can't maintain a list of
full info on each channel (such as members, etc), so the bot only maintains a
list of Names and IDs.
*/
type Channel struct {
	Name string
	ID   string
}

/*
This handler waits for the hello message and then loads the info.
*/
func (bot *Bot) helloHandler(_ *Bot, _ slack.RTMEvent) error {
	bot.Log.Info("handling hello event")
	bot.infoLock.Lock()
	info := bot.RTM.GetInfo()
	for _, user := range info.Users {
		bot.userByName[user.Name] = &user
		bot.userByID[user.ID] = &user
	}

	for _, channel := range info.Channels {
		bot.channelByName[channel.Name] = channel.ID
		bot.channelByID[channel.ID] = channel.Name
	}
	bot.infoLock.Unlock()
	return nil
}

/*
This handler listens for team_join events and adds the user to the bot's list.
*/
func (bot *Bot) teamJoinHandler(_ *Bot, evt slack.RTMEvent) error {
	join := evt.Data.(*slack.TeamJoinEvent)
	bot.Log.WithFields(logrus.Fields{
		"name": join.User.Name, "id": join.User.ID,
	}).Info("handling team_join event")
	bot.infoLock.Lock()
	bot.userByName[join.User.Name] = &join.User
	bot.userByID[join.User.ID] = &join.User
	bot.infoLock.Unlock()
	return nil
}

/*
This handler listens for user_change events and updates the user in the list.
*/
func (bot *Bot) userChangeHandler(_ *Bot, evt slack.RTMEvent) error {
	change := evt.Data.(*slack.UserChangeEvent)
	bot.Log.WithFields(logrus.Fields{
		"name": change.User.Name, "id": change.User.ID,
	}).Info("handling user_change event")
	bot.infoLock.Lock()
	bot.userByName[change.User.Name] = &change.User
	bot.userByID[change.User.ID] = &change.User
	bot.infoLock.Unlock()
	return nil
}

/*
This handler listens for channel_created events and adds the channel.
*/
func (bot *Bot) channelCreatedHandler(_ *Bot, evt slack.RTMEvent) error {
	join := evt.Data.(*slack.ChannelCreatedEvent)
	bot.Log.WithFields(logrus.Fields{
		"name": join.Channel.Name, "id": join.Channel.ID,
	}).Info("handling channel_created event")
	bot.infoLock.Lock()
	bot.channelByName[join.Channel.Name] = join.Channel.ID
	bot.channelByID[join.Channel.ID] = join.Channel.Name
	bot.infoLock.Unlock()
	return nil
}

/*
This handler listens for channel_deleted events and deletes the channel.
*/
func (bot *Bot) channelDeletedHandler(_ *Bot, evt slack.RTMEvent) error {
	del := evt.Data.(*slack.ChannelDeletedEvent)
	channelName := bot.channelByID[del.Channel]
	bot.Log.WithFields(logrus.Fields{
		"name": channelName, "id": del.Channel,
	}).Info("handling channel_deleted event")
	bot.infoLock.Lock()
	delete(bot.channelByID, del.Channel)
	delete(bot.channelByName, channelName)
	bot.infoLock.Unlock()
	return nil
}

func (bot *Bot) registerInfoHandlers() {
	bot.OnEvent("hello", bot.helloHandler)
	bot.OnEvent("team_join", bot.teamJoinHandler)
	bot.OnEvent("user_change", bot.userChangeHandler)
	bot.OnEvent("channel_created", bot.channelCreatedHandler)
	bot.OnEvent("channel_deleted", bot.channelDeletedHandler)
}

/*
Return a pointer to a User object corresponding to a User ID. Returns nil if the
user ID doesn't exist. This can be called safely from any goroutine.
*/
func (bot *Bot) GetUserByID(id string) *slack.User {
	bot.infoLock.RLock()
	user := bot.userByID[id]
	bot.infoLock.RUnlock()
	return user
}

/*
Return a pointer to a User object corresponding to a user name. Returns nil if
the user name doesn't exist. This can be called safely from any goroutine.
*/
func (bot *Bot) GetUserByName(name string) *slack.User {
	bot.infoLock.RLock()
	user := bot.userByName[name]
	bot.infoLock.RUnlock()
	return user
}

/*
Return a slice of User objects corresponding to all Users on the Slack team.
The User objects should not be modified, but the slice can be modified safely.
This function may be called safely from any goroutine.
*/
func (bot *Bot) GetUsers() []*slack.User {
	bot.infoLock.RLock()
	users := make([]*slack.User, 0, len(bot.userByID))
	for _, value := range bot.userByID {
		users = append(users, value)
	}
	bot.infoLock.RUnlock()
	return users
}

/*
Return a channel's name from its ID. Returns empty string if the channel id does
not exist. Can be called safely from any goroutine.
*/
func (bot *Bot) GetChannelByID(id string) string {
	bot.infoLock.RLock()
	channelName := bot.channelByID[id]
	bot.infoLock.RUnlock()
	return channelName
}

/*
Return a channel's ID from its name. Returns empty string if the channel name
does not exist. Can be called safely from any goroutine.
*/
func (bot *Bot) GetChannelByName(name string) string {
	bot.infoLock.RLock()
	channelID := bot.channelByName[name]
	bot.infoLock.RUnlock()
	return channelID
}

/*
Return a slice of (Name, ID) pairs corresponding to each channel in the Slack
team. The slice, and the pairs can be modified safely. This function can be
called safely from any goroutine.
*/
func (bot *Bot) GetChannels() []Channel {
	bot.infoLock.RLock()
	users := make([]Channel, 0, len(bot.channelByID))
	for id, name := range bot.channelByID {
		users = append(users, Channel{ID: id, Name: name})
	}
	bot.infoLock.RUnlock()
	return users
}
