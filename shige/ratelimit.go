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
	"time"
)

const (
	period           = 30
	userMessageLimit = 19
	modMessageLimit  = 99
)

type rateLimiter struct {
	messageCounter        int
	lastMessageCountReset int64
	messageLimit          int
}

func (b *Bot) initRateLimiter() {
	rl := &rateLimiter{0, 0, userMessageLimit}
	if b.isMod {
		rl.messageLimit = modMessageLimit
	}
	b.rateLimiter = rl
	fmt.Printf("> Initialized rate limiter with msglimit=%d\n", rl.messageLimit)
}

// Privmsgf formats and sends a rate-limited message to channel.
func (b *Bot) Privmsgf(channel, format string, args ...interface{}) {
	now := time.Now().Unix()
	b.w.Await(func() {
		// TODO: modify this to block for the least time possible
		rl := b.rateLimiter
		if rl.lastMessageCountReset == 0 ||
			now-rl.lastMessageCountReset > period {

			fmt.Printf("> Rate limiter: %d messages sent in %d seconds "+
				"(limit is %d msgs / %d seconds).\n",
				rl.messageCounter, now-rl.lastMessageCountReset,
				rl.messageLimit, period)

			rl.lastMessageCountReset = now
			rl.messageCounter = 0
		}

		if rl.messageCounter >= rl.messageLimit {
			go func() {
				difference := now - rl.lastMessageCountReset
				amount := time.Second*period -
					time.Duration(int64(time.Second)*difference) +
					time.Millisecond*500

				fmt.Println("!! Rate limit reached, postponing this message by",
					amount)
				<-time.After(amount)
				fmt.Println("Sending delayed message")
				b.Privmsgf(channel, format, args...)
			}()
			return
		}

		b.irc.Privmsgf(channel, fmt.Sprintf("%s %s", format, b.randStr()), args...)
		rl.messageCounter++
	})
}
