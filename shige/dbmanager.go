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
	_ "github.com/cznic/ql/driver"
	//_ "github.com/mattn/go-sqlite3"
	"os"
)

const commandsFile = "shige_ql.db"

type dbManager struct{ *sql.DB }

func (b *Bot) initDB() (err error) {
	b.db, err = newDBManager()
	//b.db.convertDB()
	return
}

/*
func (db dbManager) convertDB()  {
	fmt.Println("CONVERTINGU CONVERTINGU")
	conn, err := sql.Open("sqlite3", "shige.db")
	if err != nil {
		return
	}

	olddb := dbManager{conn}
	
	sqlStmt, err := olddb.Prepare(
		"select channel, name, reply, mod_only from commands")
	if err != nil {
		panic(err)
	}

	rows, err := sqlStmt.Query()
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var channel, name, reply string
		var modOnly bool
		err = rows.Scan(&channel, &name, &reply, &modOnly)
		if err != nil {
			panic(err)
		}
		db.setCommand(channel, name, reply, modOnly)
	}
	
	rows.Close()
	sqlStmt.Close()
	
	sqlStmt, err = olddb.Prepare(
		"select channel, url from gists")
	if err != nil {
		panic(err)
	}

	rows, err = sqlStmt.Query()
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var channel, url string
		err = rows.Scan(&channel, &url)
		if err != nil {
			panic(err)
		}
		db.setGist(channel, url)
	}
	
	rows.Close()
	sqlStmt.Close()
}
*/

func newDBManager() (db dbManager, err error) {
	_, err = os.Stat(commandsFile)
	createTables := os.IsNotExist(err)

	conn, err := sql.Open("ql", commandsFile)
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
		channel string not null, 
		name string not null, 
		reply string not null, 
		mod_only bool not null
	);
	create table gists (
		channel string not null, 
		url string not null
	);
	create unique index commands_index on commands(channel, name);
	create unique index gists_index on gists(channel);`
	
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()
	
	_, err = tx.Exec(sqlStmt)
	return
}

func (db dbManager) getGist(channel string) (gistUrl string) {
	fmt.Println("DB: Getting gist for", channel)
	sqlStmt, err := db.Prepare("select url from gists where channel==$1;")
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
		sqlStmt, err := tx.Prepare("update gists set url=$1 where channel==$2;")
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
	sqlStmt, err := tx.Prepare(
		"insert into gists(channel, url) values($1, $2);")
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
		"select reply, mod_only from commands where channel==$1 and name==$2;")
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

	err = rows.Scan(&text, &modOnly)
	if err != nil {
		panic(err)
	}

	return
}

func (db dbManager) getCommands(channel string) (res map[string]*TextCommand) {
	fmt.Println("DB: Loading commands for", channel)
	res = make(map[string]*TextCommand)

	sqlStmt, err := db.Prepare(
		"select name, reply, mod_only from commands where channel==$1;")
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
		err = rows.Scan(&name, &c.Text, &c.ModOnly)
		if err != nil {
			panic(err)
		}
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
		sqlStmt, err := tx.Prepare("update commands set reply=$1, mod_only=$2" +
			" where channel==$3 and name==$4;")
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
			"values($1, $2, $3, $4);")
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
		"delete from commands where channel==$1 and name==$2;")
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
