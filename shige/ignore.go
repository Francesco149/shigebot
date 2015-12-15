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

func (b *Bot) initIgnoreList(botnick string) {
	b.ignore = make(map[string]bool)
	b.ignore[botnick] = true
}

// Ignore ignores text commands for a list of nicks.
// Note: the bot ignores itself by default.
func (b Bot) Ignore(nicknames ...string) {
	b.w.Await(func() {
		for _, nick := range nicknames {
			b.ignore[nick] = true
		}
	})
}

// Unignore restores text commands for a list of nicks.
func (b Bot) Unignore(nicknames ...string) {
	b.w.Await(func() {
		for _, nick := range nicknames {
			delete(b.ignore, nick)
		}
	})
}

// Ignored returns whether text commands are ignored for the nickname.
func (b Bot) Ignored(nick string) bool {
	resp := make(chan bool, 1)
	b.w.Do(func() {
		resp <- b.ignore[nick]
		close(resp)
	})
	return <-resp
}
