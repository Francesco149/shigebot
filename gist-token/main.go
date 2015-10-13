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
	"bufio"
	"fmt"
	"github.com/MaximeD/gost/conf/oauth"
	"os"
)

func main() {
	fmt.Println("This is your OAuth gist token, please copy it to your",
		"config.json:", OAuth.GetToken())
	fmt.Println("If you accidentally lose it, you can always see your tokens",
		"from here: https://github.com/settings/applications and edit and",
		"regenerate the gost token")
	fmt.Print("Press 'Enter' to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
