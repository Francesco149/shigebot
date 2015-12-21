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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
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
func (b *Bot) AddCommand(name string, handler func(*CommandData)) {
	b.w.Await(func() { b.commands[name] = handler })
}

// RemoveCommand removes a command.
func (b *Bot) RemoveCommand(name string) {
	b.w.Await(func() { delete(b.commands, name) })
}

// CommandExists returns whether the command exists.
func (b *Bot) CommandExists(name string) bool {
	resp := make(chan bool, 1)
	b.w.Do(func() {
		resp <- b.commands[name] != nil
		close(resp)
	})
	return <-resp
}

// Command returns a command's handler.
func (b *Bot) Command(name string) func(*CommandData) {
	resp := make(chan func(*CommandData), 1)
	b.w.Do(func() {
		resp <- b.commands[name]
		close(resp)
	})
	return <-resp
}

func (b *Bot) isCooldown(ch *Channel, command string) bool {
	resp := make(chan bool, 1)
	b.w.Do(func() {
		res := time.Since(ch.builtinLastUsage[command]) <
			time.Duration(ch.commandCooldown)*time.Millisecond
		if !res {
			ch.builtinLastUsage[command] = time.Now()
		}
		resp <- res
		close(resp)
	})
	return <-resp
}

func (b *Bot) initCommands() {
	// TODO: join builtin commands with the channel commands somehow
	b.commands = map[string]func(*CommandData){
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

			if !b.caseSensitive {
				commandName = strings.ToLower(commandName)
			}

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
			if !b.caseSensitive {
				commandName = strings.ToLower(commandName)
			}
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
			if !b.caseSensitive {
				commandName = strings.ToLower(commandName)
			}
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
			if !b.caseSensitive {
				commandName = strings.ToLower(commandName)
			}
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

			resp := make(chan int32, 1)
			b.w.Do(func() {
				resp <- ch.commandCooldown
				close(resp)
			})
			cd := <-resp

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

			b.w.Await(func() { ch.commandCooldown = int32(i) })
			ch.Privmsgf("Command cooldown set to %v milliseconds", i)
		},

		"uptime": func(c *CommandData) {
			ch := c.Channel

			if b.isCooldown(ch, "uptime") {
				fmt.Println("uptime is on cooldown")
				return
			}

			req, err := http.NewRequest("GET",
				"https://api.twitch.tv/kraken/streams/"+ch.name[1:],
				nil)
			if err != nil {
				ch.Privmsgf("API error: %v", err)
				return
			}

			fmt.Println(req.URL)

			req.Header.Add("Accept", "application/vnd.twitchtv.v3+json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				ch.Privmsgf("API error: %v", err)
				return
			}

			defer resp.Body.Close()

			var res map[string]interface{}

			err = json.NewDecoder(resp.Body).Decode(&res)
			if err != nil {
				ch.Privmsgf("API error: %v", err)
				return
			}

			// lol too lazy to do proper json decoding with structs
			stream, ok := res["stream"].(map[string]interface{})
			createdAt, ok := stream["created_at"].(string)
			if len(createdAt) == 0 || !ok {
				ch.Privmsgf("Offline")
				return
			}

			parsedTime, err := time.Parse(
				"2006-01-02T15:04:05Z", createdAt)
			if err != nil {
				ch.Privmsgf("API error: %v", err)
				return
			}

			ch.Privmsgf("%v", time.Now().UTC().Sub(parsedTime))
		},
	}

	b.BuiltinCommandsInfo = "" +
		`* +!cmdadd: adds a command (Usage: !cmdadd commandname text)
* +!cmdremove: removes a command (Usage: !cmdremove commandname)
* +!cmdedit: changes the text for a command (Usage: !cmdedit commandname text)
* +!modonly: limits a command to mods only (Usage: !modonly commandname yes/no)
* +!cooldown: milliseconds before a command can be reused (Usage: !cooldown ms)
* !uptime: shows the channel's uptime if online`

	fmt.Println("> Built-in commands initialized")
}
