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
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
)

// I'm aware all of these could be methods for Channel but I prefer keeping the
// built-in commands outside of the channel entity.

// CommandData holds information about a chat message that contains a recognized
// command.
type CommandData struct {
	// Channel is a pointer to the channel where this message originated.
	Channel *Channel
	// Args contains every parameter after the command name. It is built using
	// strings.Fields, so repeated whitespace is ignored.
	Args []string
	// Nick is the nickname of the user that sent the command.
	Nick string
}

// strips ! from a command name if present.
func parseCommandName(str string) string {
	if str[0] == '!' {
		str = str[1:]
	}
	return str
}

// AddCommand adds a command and binds it to handler.
func (b Bot) AddCommand(name string, handler func(*CommandData)) {
	commands := <-b.chCommands
	commands[name] = handler
	b.chCommands <- commands
}

// RemoveCommand removes a command.
func (b Bot) RemoveCommand(name string) {
	commands := <-b.chCommands
	delete(commands, name)
	b.chCommands <- commands
}

// CommandExists returns whether the command exists.
func (b Bot) CommandExists(name string) bool {
	commands := <-b.chCommands
	defer func() { b.chCommands <- commands }()
	return commands[name] != nil
}

func (b *Bot) initCommands() {
	commands := map[string]func(*CommandData){
		"cmdadd": func(c *CommandData) {
			ch := c.Channel
			if !ch.IsMod(c.Nick) {
				return
			}
			if len(c.Args) < 2 {
				ch.Privmsgf("Usage: !cmdadd commandname text")
				return
			}

			// first argument is command name, all the remaining
			// arguments combine to the command text
			commandName := parseCommandName(c.Args[0])
			commandText := strings.Join(c.Args[1:], " ")
			err := ch.AddCommand(commandName, commandText)
			if err != nil {
				ch.Privmsgf("%v", err)
				return
			}

			ch.Privmsgf("Added command %s", commandName)
			b.updateCommandList(c.Channel)
		},

		"cmdremove": func(c *CommandData) {
			ch := c.Channel
			if !ch.IsMod(c.Nick) {
				return
			}
			if len(c.Args) != 1 {
				ch.Privmsgf("Usage: !cmdremove commandname")
				return
			}

			commandName := parseCommandName(c.Args[0])
			if b.CommandExists(commandName) {
				ch.Privmsgf("Command %s cannot be removed.", commandName)
				return
			}

			err := ch.RemoveCommand(commandName)
			if err != nil {
				ch.Privmsgf("%v", err)
				return
			}

			ch.Privmsgf("Removed command %s", commandName)
			b.updateCommandList(c.Channel)
		},

		"cmdedit": func(c *CommandData) {
			ch := c.Channel
			if !ch.IsMod(c.Nick) {
				return
			}
			if len(c.Args) < 2 {
				ch.Privmsgf("Usage: !cmdedit commandname text")
				return
			}

			commandName := parseCommandName(c.Args[0])
			if b.CommandExists(commandName) {
				ch.Privmsgf("Command %s cannot be edited.", commandName)
				return
			}

			commandText := strings.Join(c.Args[1:], " ")
			err := ch.EditCommand(commandName, commandText)
			if err != nil {
				ch.Privmsgf("%v", err)
				return
			}

			ch.Privmsgf("Edited command %s", commandName)
			b.updateCommandList(c.Channel)
		},

		"modonly": func(c *CommandData) {
			ch := c.Channel
			if !ch.IsMod(c.Nick) {
				return
			}
			if len(c.Args) != 2 || (c.Args[1] != "yes" && c.Args[1] != "no") {
				ch.Privmsgf("Usage: !modonly commandname yes/no")
				return
			}

			commandName := parseCommandName(c.Args[0])
			if b.CommandExists(commandName) {
				ch.Privmsgf("Command %s cannot be edited.", commandName)
				return
			}

			toggle := c.Args[1] == "yes"
			err := ch.SetCommandMod(commandName, toggle)
			if err != nil {
				ch.Privmsgf("%v", err)
				return
			}

			ch.Privmsgf("Command %s modonly = %v.", commandName, toggle)
			b.updateCommandList(c.Channel)
		},

		"cooldown": func(c *CommandData) {
			ch := c.Channel
			if !ch.IsMod(c.Nick) {
				return
			}

			cd := atomic.LoadInt32(&ch.commandCooldown)
			usage := fmt.Sprintf(
				"Usage: !cooldown milliseconds. Current cooldown is %vms.", cd)

			if len(c.Args) != 1 {
				ch.Privmsgf(usage)
				return
			}

			i, err := strconv.ParseInt(c.Args[0], 10, 32)
			if err != nil {
				ch.Privmsgf(usage)
				return
			}

			if i < 0 {
				i = 0
			}

			ch.Privmsgf("Setting command cooldown to %v milliseconds", i)
			atomic.StoreInt32(&ch.commandCooldown, int32(i))
		},
	}

	b.chCommands = make(chan map[string]func(*CommandData), 1)
	b.chCommands <- commands

	b.BuiltinCommandsInfo = "" +
		`* +!cmdadd: adds a command (Usage: !cmdadd commandname text)
* +!cmdremove: removes a command (Usage: !cmdremove commandname)
* +!cmdedit: changes the text for a command (Usage: !cmdedit commandname text)
* +!modonly: limits a command to mods only (Usage: !modonly commandname yes/no)
* +!cooldown: milliseconds before a command can be reused (Usage: !cooldown ms)`

	fmt.Println("> Built-in commands initialized")
}
