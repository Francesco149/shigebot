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
	"math/rand"
)

func (b Bot) randStr() string {
	if b.isMod {
		return ""
	}
	return fmt.Sprintf("[%d]", rand.Int31n(99))
}

func attemptQuery(q func() error) error {
	var err error
	for i := 0; i < 5; i++ {
		// make 5 attempts just in case concurrent queries made the query fail
		err = q()
		if err != nil {
			fmt.Println(err)
			continue
		}
		break
	}
	if err != nil {
		return errors.New("Database error. Please try again.")
	}
	return nil
}
