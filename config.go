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
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type config struct {
	GistOAuth     string
	TwitchUser    string
	TwitchOAuth   string
	Ignore        []string
	Channels      []string
	IsMod         bool
	CaseSensitive bool
}

func loadConfig() (conf *config, err error) {
	jsonBlob, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println("Failed to load config", err)
		return
	}

	conf = &config{}
	err = json.Unmarshal(jsonBlob, conf)
	if err != nil {
		fmt.Println("Failed to parse config", err)
		return
	}

	return
}
