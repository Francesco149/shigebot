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

package shige

import (
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
	"time"
)

// A Channel is a single irc channel to which the bot is connected.
type Channel struct {
	name            string
	chMods          chan map[string]bool
	chCommands      chan map[string]*TextCommand
	parent          *Bot
	commandCooldown int64
}

// I don't really need a map for mods but looking up names is less code.
// chMods and chCommands need to be thread safe as the irc library fires events
// asynchronously as goroutines.

func newChannel(parent *Bot, name string) *Channel {
	c := &Channel{
		name,
		make(chan map[string]bool, 1),
		make(chan map[string]*TextCommand, 1),
		parent,
		0,
	}
	c.chMods <- make(map[string]bool)

	// load commands for this channel
	commands := parent.db.getCommands(name)
	c.chCommands <- commands

	addhelp := func() {
		if c.CommandExists("help") {
			return
		}
		c.AddCommand("help", fmt.Sprintf("Command list: %s",
			parent.db.getGist(c.name)))
	}

	// if the gist is already initialized, add help now and update the gist
	gistInitialized := parent.db.getGist(c.name) != ""

	if gistInitialized {
		addhelp()
	}

	// refresh gist
	parent.updateCommandList(c)

	// if the gist wasn't initialized earlier, add help now and update the gist
	if !gistInitialized {
		addhelp()
		parent.updateCommandList(c)
	}

	return c
}

func (c Channel) filename() string {
	return fmt.Sprintf("%s.txt", c.name[1:])
}

// Printf calls fmt.Printf with the channel name as a prefix to the text.
func (c Channel) Printf(format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("%s> %s", c.name, format), args...)
}

// Println calls fmt.Println with the channel name as a prefix to the text.
func (c Channel) Println(args ...interface{}) {
	fmt.Println(append([]interface{}{fmt.Sprintf("%s>", c.name)}, args...)...)
}

// Privmsgf formats and sends a rate-limited message.
func (c Channel) Privmsgf(format string, args ...interface{}) {
	c.parent.Privmsgf(c.name, format, args...)
}

// AddMod allows nick to use mod commands.
func (c *Channel) AddMod(nick string) {
	c.Println("Adding mod", nick)
	mods := <-c.chMods
	mods[nick] = true
	c.chMods <- mods
}

// RemoveMod revokes the use of mod commands for nick.
func (c *Channel) RemoveMod(nick string) {
	c.Println("Removing mod", nick)
	mods := <-c.chMods
	delete(mods, nick)
	c.chMods <- mods
}

// IsMod returns whether nick is allowed to use mod commands.
func (c Channel) IsMod(nick string) bool {
	mods := <-c.chMods
	defer func() { c.chMods <- mods }()
	return mods[nick]
}

// FullCommandList retrieves a list of the commands and their description (or
// text if they are simple text commands) separated by the string separator.
// Mod commands will be prefixed by the modPrefix string.
// If noDescription is true, description or text will be omitted.
// The list is alphabetically sorted.
func (c Channel) FullCommandList(separator, modPrefix string,
	noDescription bool) (res string) {

	commands := <-c.chCommands
	defer func() { c.chCommands <- commands }()

	// sort commands by name
	sorted := make([]string, 0)
	for name := range commands {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)

	for _, name := range sorted {
		command := commands[name]
		line := ""
		if noDescription {
			line = fmt.Sprintf("!%s%s", name, separator)
		} else {
			line = fmt.Sprintf("!%s: %s%s", name, command.Text, separator)
		}
		if command.ModOnly {
			line = fmt.Sprintf("%s%s", modPrefix, line)
		}
		res += line
	}

	// remove last trailing separator
	if len(sorted) > 0 {
		res = res[:len(res)-len(separator)]
	}
	return
}

// CommandList returns a comma-separated list of the commands, prefixing
// moderator commands with a +.
// The list is alphabetically sorted.
func (c Channel) CommandList() string {
	return c.FullCommandList(", ", "+", true)
}

// Command returns a pointer to a command.
func (c Channel) Command(name string) *TextCommand {
	commands := <-c.chCommands
	defer func() { c.chCommands <- commands }()
	return commands[name]
}

// AddCommand adds a simple text command.
func (c Channel) AddCommand(name, text string) error {
	commands := <-c.chCommands
	defer func() { c.chCommands <- commands }()
	if commands[name] != nil {
		return errors.New(fmt.Sprintf("Command %s already exists.", name))
	}

	err := attemptQuery(func() error {
		return c.parent.db.setCommand(c.name, name, text, false)
	})
	if err != nil {
		return err
	}

	commands[name] = &TextCommand{Text: text, ModOnly: false}
	c.Println("Added command", name, "->", text)
	return nil
}

// RemoveCommand removes a simple text command.
func (c Channel) RemoveCommand(name string) error {
	commands := <-c.chCommands
	defer func() { c.chCommands <- commands }()
	if commands[name] == nil {
		return errors.New(fmt.Sprintf("Command %s doesn't exist.", name))
	}

	err := attemptQuery(func() error {
		return c.parent.db.removeCommand(c.name, name)
	})
	if err != nil {
		return err
	}

	delete(commands, name)
	c.Println("Removed command", name)
	return nil
}

// EditCommand replaces the text of an existing simple text command.
func (c Channel) EditCommand(name, text string) error {
	commands := <-c.chCommands
	defer func() { c.chCommands <- commands }()
	if commands[name] == nil {
		return errors.New(fmt.Sprintf("Command %s doesn't exist.", name))
	}

	err := attemptQuery(func() error {
		return c.parent.db.setCommand(c.name, name, text, false)
	})
	if err != nil {
		return err
	}

	commands[name].Text = text
	c.Println("Edited command", name, "->", text)
	return nil
}

// SetCommandMod sets whehter a command is for mods only or not.
func (c Channel) SetCommandMod(name string, modOnly bool) error {
	commands := <-c.chCommands
	defer func() { c.chCommands <- commands }()
	if commands[name] == nil {
		return errors.New(fmt.Sprintf("Command %s doesn't exist.", name))
	}

	err := attemptQuery(func() error {
		return c.parent.db.setCommand(
			c.name, name, commands[name].Text, modOnly)
	})
	if err != nil {
		return err
	}

	commands[name].ModOnly = modOnly
	return nil
}

// CommandExists returns whether a command exists.
func (c Channel) CommandExists(name string) bool {
	return c.Command(name) != nil
}

func (c Channel) onCommand(commandName, nick string) bool {
	if !c.CommandExists(commandName) {
		// should only return false when the command doesn't exist
		return false
	}

	commands := <-c.chCommands
	defer func() { c.chCommands <- commands }()
	command := commands[commandName]

	cd := atomic.LoadInt64(&c.commandCooldown)
	elapsed := time.Now().Sub(command.LastUsage)
	cooldown := time.Millisecond * time.Duration(cd)
	if elapsed < cooldown {
		c.Println("Rejected command", commandName,
			"because it is still on cooldown,", elapsed,
			"since last usage, cooldown is", cooldown)
		return true
	}

	if command.ModOnly && !c.IsMod(nick) {
		c.Println("Rejected command", commandName,
			"because the user is not a mod")
		return true
	}

	command.LastUsage = time.Now()
	c.Println("Processing text command", commandName)
	c.Privmsgf("%s", command.Text)
	return true
}
