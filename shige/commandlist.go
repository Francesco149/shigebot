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
	"github.com/Francesco149/shigebot/shige/gist"
	"io/ioutil"
	"strings"
)

const (
	githubApi = "https://api.github.com/"
	gistDesc  = "Shigebot Commands for "
)

func (b *Bot) updateCommandList(ch *Channel) {
	channel := ch.name
	commands := fmt.Sprintf(
		`# %s
by Franc\[e\]sco / lolisamurai

Available commands for channel [%s](http://www.twitch.tv/%s) (+ = mod only):

%s
`,
		BotName, channel, channel[1:], b.BuiltinCommandsInfo)

	list := ch.FullCommandList("\n", "+", false)
	lines := strings.Split(list, "\n")
	for _, line := range lines {
		commands += "* " + line + "\n"
	}

	filename := fmt.Sprintf("commands-for-%s.md", channel[1:])
	err := ioutil.WriteFile(filename, []byte(commands), 0777)
	if err != nil {
		panic(err)
	}

	for {
		if !b.db.gistExists(channel) {
			url, err := gist.Post(githubApi, b.gistOAuth, true,
				[]string{filename}, gistDesc+channel)
			if err != nil {
				fmt.Println(err)
				break
			}
			err = attemptQuery(func() error {
				return b.db.setGist(channel, url)
			})
		} else {
			url := b.db.getGist(channel)
			err = gist.Update(githubApi, b.gistOAuth, []string{filename},
				url, gistDesc+channel)
		}

		break
	}

	if err != nil {
		fmt.Println("Failed to update command list gist, will retry next time!")
	}
}
