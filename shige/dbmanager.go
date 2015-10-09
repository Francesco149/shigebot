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
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

const commandsFile = "shige.db"

type dbManager struct{ *sql.DB }

func (b *Bot) initDB() (err error) {
	b.db, err = newDBManager()
	return
}

func newDBManager() (db dbManager, err error) {
	_, err = os.Stat(commandsFile)
	createTables := os.IsNotExist(err)

	conn, err := sql.Open("sqlite3", commandsFile)
	if err != nil {
		return
	}

	db = dbManager{conn}

	if !createTables {
		return
	}

	err = nil

	fmt.Println("DB: Initializing tables")
	sqlStmt := `
	create table commands (
		channel char(512) not null, 
		name char(512) not null, 
		reply text not null, 
		mod_only int not null, 
		primary key (channel, name)
	);
	create table gists (
		channel char(512) not null, 
		url text not null, 
		primary key (channel)
	);
	`
	_, err = db.Exec(sqlStmt)
	return
}

func (db dbManager) getGist(channel string) (gistUrl string) {
	fmt.Println("DB: Getting gist for", channel)
	sqlStmt, err := db.Prepare("select url from gists where channel=?")
	if err != nil {
		panic(err)
	}
	defer sqlStmt.Close()

	rows, err := sqlStmt.Query(channel)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	if !rows.Next() {
		return
	}

	err = rows.Scan(&gistUrl)
	if err != nil {
		panic(err)
		return
	}

	return
}

func (db dbManager) gistExists(channel string) bool {
	return len(db.getGist(channel)) != 0
}

func (db dbManager) setGist(channel, gistUrl string) error {
	justUpdate := db.gistExists(channel)

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	if justUpdate {
		fmt.Println("DB: Updating gist for", channel)
		sqlStmt, err := tx.Prepare("update gists set url=? where channel=?")
		if err != nil {
			panic(err)
		}
		defer sqlStmt.Close()

		_, err = sqlStmt.Exec(gistUrl, channel)
		if err != nil {
			return err
		}
		return nil
	}

	fmt.Println("DB: Adding gist for", channel)
	sqlStmt, err := tx.Prepare("insert into gists(channel, url) values(?, ?);")
	if err != nil {
		panic(err)
	}
	defer sqlStmt.Close()

	_, err = sqlStmt.Exec(channel, gistUrl)
	if err != nil {
		return err
	}

	return nil
}

func (db dbManager) getCommand(channel, command string) (
	text string, modOnly bool) {

	fmt.Println("DB: Getting command", command, "in", channel)
	sqlStmt, err := db.Prepare(
		"select reply, mod_only from commands where channel=? and name=?")
	if err != nil {
		panic(err)
	}
	defer sqlStmt.Close()

	rows, err := sqlStmt.Query(channel, command)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	if !rows.Next() {
		return
	}

	var tmpModOnly int
	err = rows.Scan(&text, &tmpModOnly)
	if err != nil {
		panic(err)
	}
	modOnly = tmpModOnly == 1

	return
}

func (db dbManager) getCommands(channel string) (res map[string]*TextCommand) {
	fmt.Println("DB: Loading commands for", channel)
	res = make(map[string]*TextCommand)

	sqlStmt, err := db.Prepare(
		"select name, reply, mod_only from commands where channel=?")
	if err != nil {
		panic(err)
	}
	defer sqlStmt.Close()

	rows, err := sqlStmt.Query(channel)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		c := &TextCommand{}
		var name string
		var tmpModOnly int
		err = rows.Scan(&name, &c.Text, &tmpModOnly)
		if err != nil {
			panic(err)
		}
		c.ModOnly = tmpModOnly == 1
		res[name] = c
		fmt.Println(res[name])
	}

	return
}

func (db dbManager) commandExists(channel, command string) bool {
	text, _ := db.getCommand(channel, command)
	return len(text) != 0
}

func (db dbManager) setCommand(channel, command,
	text string, modOnly bool) error {

	justUpdate := db.commandExists(channel, command)

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	if justUpdate {
		fmt.Println("DB: Updating command", command, "for", channel)
		sqlStmt, err := tx.Prepare("update commands set reply=?, mod_only=? " +
			"where channel=? and name=?")
		if err != nil {
			panic(err)
		}
		defer sqlStmt.Close()

		_, err = sqlStmt.Exec(text, modOnly, channel, command)
		if err != nil {
			return err
		}
		return nil
	}

	fmt.Println("DB: Adding command", command, "for", channel)
	sqlStmt, err := tx.Prepare(
		"insert into commands(channel, name, reply, mod_only) " +
			"values(?, ?, ?, ?);")
	if err != nil {
		panic(err)
	}
	defer sqlStmt.Close()

	_, err = sqlStmt.Exec(channel, command, text, modOnly)
	if err != nil {
		return err
	}

	return nil
}

func (db dbManager) removeCommand(channel, command string) error {
	fmt.Println("DB: Removing command", command, "for", channel)
	if !db.commandExists(channel, command) {
		fmt.Println("DB:", command, "doesn't exist, so no need to remove it")
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	sqlStmt, err := tx.Prepare(
		"delete from commands where channel=? and name=?")
	if err != nil {
		panic(err)
	}
	defer sqlStmt.Close()

	_, err = sqlStmt.Exec(channel, command)
	if err != nil {
		return err
	}

	return nil
}
