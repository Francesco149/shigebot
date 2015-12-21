/*
	Copyright 2015 Franc[e]sco (lolisamurai@tfwno.gf)
	This file is part of Shigebot.
	Shigebot is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	Shigebot is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with Shigebot. If not, see <http://www.gnu.org/licenses/>.
*/

// Package shige implements Shigebot, a twitch irc bot.
package shige

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"strings"
)

const BotName = "Shigebot 1.2.2"

// A Bot is an instance of Shigebot connected to multiple channels on twitch
// on a single twitch account.
type Bot struct {
	// If not nil, this function will be called when a PRIVMSG is received.
	// The function must return true if the handling of this message should
	// continue (the message will be forwarded to all the internal command
	// handlers) or false otherwise.
	OnPrivmsg func(*irc.Event) bool

	// The documentation for all the built-in commands. This is already set
	// by default, but it can be modified when there is a need to customize the
	// built-in commands or how they show up in the gist.
	BuiltinCommandsInfo string

	irc           *irc.Connection
	w             *Worker
	db            dbManager
	isMod         bool
	caseSensitive bool
	gistOAuth     string
	channels      map[string]*Channel
	commands      map[string]func(*CommandData)
	rateLimiter   *rateLimiter
	ignore        map[string]bool
}

// Irc returns a pointer to the irc connection object used by the bot.
func (b Bot) Irc() *irc.Connection { return b.irc }

// Channel returns a pointer to a channel.
func (b Bot) Channel(channel string) *Channel {
	resp := make(chan *Channel, 1)
	b.w.Do(func() {
		resp <- b.channels[channel]
		close(resp)
	})
	return <-resp
}

// Join makes the bot join channel and load any commands that might have been
// previously saved for that channel.
func (b *Bot) Join(channel string) {
	fmt.Println("> Joining", channel)
	ch := newChannel(b, channel)
	b.w.Await(func() {
		b.irc.Join(channel)
		b.channels[channel] = ch
	})
}

// Part makes the bot leave channel.
func (b Bot) Part(channel string) {
	fmt.Println("> Leaving", channel)
	b.irc.Part(channel)
	b.w.Await(func() { delete(b.channels, channel) })
}

// Init initializes a new instance of Bot and connects to twitch irc servers
// using twitchUser and twitchOauth as credentials. Each channel in channelList
// is joined (channel names must include the # prefix).
// The isMod flag specifies whether the bot's account is a moderator in the
// channels it will join. Running in non-moderator mode will result in a lower
// message rate limit as well as randomization of each message by appending
// a random number to bypass twitch spam prevention.
// caseSensitive makes text commands case sensitive if true.
// gistOAuth is the github oauth token that will be used to upload the command
// list.
// Returns a pointer to the Bot instance and an error if anything goes wrong.
func Init(twitchUser, twitchOauth, gistOAuth string, channelList []string,
	isMod, caseSensitive bool) (b *Bot, err error) {

	fmt.Printf("> %s\n", BotName)
	b = &Bot{
		isMod:         isMod,
		caseSensitive: caseSensitive,
		gistOAuth:     gistOAuth,
		w:             NewWorker("shigebot", 500),
	}

	// initialize everything
	err = b.initDB()
	if err != nil {
		return
	}
	b.initCommands()
	b.initRateLimiter()
	b.initIgnoreList(twitchUser)

	// connect to twitch irc
	ircobj := irc.IRC(twitchUser, twitchUser)
	ircobj.Password = twitchOauth

	b.irc = ircobj

	err = ircobj.Connect("irc.twitch.tv:6667")
	if err != nil {
		return
	}

	b.w.Start()
	// irc callbacks
	ircobj.AddCallback("001", func(e *irc.Event) {
		ircobj.SendRaw("CAP REQ :twitch.tv/membership") // userlist & modesets

		// join all channels
		b.w.Await(func() { b.channels = make(map[string]*Channel) })
		for _, channel := range channelList {
			b.Join(channel)
		}
	})

	ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {
		if b.OnPrivmsg != nil && !b.OnPrivmsg(event) {
			return
		}

		msg := event.Message()

		channelName := event.Arguments[0]
		nick := event.Nick

		c := b.Channel(channelName)
		c.Printf("%s: %s\n", nick, msg)

		// ignore empty messages
		if len(msg) <= 1 {
			return
		}

		// only handle commands
		if msg[0] != '!' {
			return
		}

		cmd := msg[1:]               // strip the ! off the command name
		split := strings.Fields(cmd) // split at whitespace
		cmd = split[0]
		args := split[1:]

		if !b.caseSensitive {
			cmd = strings.ToLower(cmd)
		}

		if len(cmd) == 0 {
			c.Println("Ignored empty command")
			return
		}

		builtinCommand := b.Command(cmd)

		switch {
		// global built-in commands
		case builtinCommand != nil:
			c.Println("Processing command", cmd, args)
			builtinCommand(&CommandData{c, args, nick})

		// simple text commands
		// args are not used here.
		case !b.Ignored(nick) && c.onCommand(cmd, nick):
			break

		// if OnCommand did not recognize the command, then it's definitely
		// an invalid one
		default:
			c.Println("Invalid command", cmd)
		}
	})

	ircobj.AddCallback("MODE", func(event *irc.Event) {
		fmt.Println("MODE", event.Arguments)

		// we only want user modesets
		if len(event.Arguments) != 3 {
			return
		}

		ch := event.Arguments[0]
		mode := event.Arguments[1][1:]
		operation := event.Arguments[1][:1]
		u := event.Arguments[2]

		// not setting operators
		if mode != "o" {
			return
		}

		c := b.Channel(ch)

		// add/remove operator
		switch operation {
		case "+":
			c.AddMod(u)
			break
		case "-":
			c.RemoveMod(u)
			break
		}
	})

	return
}

// Run starts the bot, allowing it to start handling commands.
func (b Bot) Run() {
	b.irc.Loop()
	fmt.Println("Waiting for worker to terminate")
	b.w.Terminate()
	return
}
