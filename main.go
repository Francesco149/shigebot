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

package main

import (
	"fmt"
	"github.com/Francesco149/shigebot/shige"
)

func main() {
	conf, err := loadConfig()
	if err != nil {
		return
	}

	bot, err := shige.Init(conf.TwitchUser, conf.TwitchOAuth, conf.GistOAuth,
		conf.Channels, conf.IsMod, conf.CaseSensitive)
	if err != nil {
		fmt.Println("Failed to initialize bot", err)
		return
	}

	bot.Ignore(conf.Ignore...)
	bot.Run()
}
