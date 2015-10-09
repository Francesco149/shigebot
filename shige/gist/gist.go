// I just took https://github.com/MaximeD/gost/blob/master/gist/gist.go
// and modified it to suit my bot, so credits to MaximeD.

// Package gist implements utilities to communicate with the github gist API.
package gist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MaximeD/gost/json"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

// Post posts a list of files to gist.
func Post(baseUrl string, accessToken string, isPublic bool,
	filesPath []string, description string) (url string, err error) {

	files := make(map[string]GistJSON.File)

	for i := 0; i < len(filesPath); i++ {
		var content []byte
		content, err = ioutil.ReadFile(filesPath[i])
		if err != nil {
			return
		}
		fileName := path.Base(filesPath[i])
		files[fileName] = GistJSON.File{Content: string(content)}
	}

	gist := GistJSON.Post{Desc: description, Public: isPublic, Files: files}

	// encode json
	buf, err := json.Marshal(gist)
	if err != nil {
		return
	}
	jsonBody := bytes.NewBuffer(buf)

	// post json
	postUrl := baseUrl + "gists"
	if accessToken != "" {
		postUrl = postUrl + "?access_token=" + accessToken
	}

	resp, err := http.Post(postUrl, "text/json", jsonBody)
	if err != nil {
		return
	}

	// close connection
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var jsonRes GistJSON.Response
	err = json.Unmarshal(body, &jsonRes)
	if err != nil {
		return
	}

	// display result
	url = jsonRes.HtmlUrl
	return
}

// Update updates an existing gist.
func Update(baseUrl string, accessToken string, filesPath []string,
	gistUrl string, description string) (err error) {

	files := make(map[string]GistJSON.File)

	for i := 0; i < len(filesPath); i++ {
		var content []byte
		content, err = ioutil.ReadFile(filesPath[i])
		if err != nil {
			return
		}
		fileName := path.Base(filesPath[i])
		files[fileName] = GistJSON.File{Content: string(content)}
	}

	gist := GistJSON.Patch{Desc: description, Files: files}

	// encode json
	buf, err := json.Marshal(gist)
	if err != nil {
		return
	}
	jsonBody := bytes.NewBuffer(buf)

	gistId := getGistId(gistUrl)

	// post json
	postUrl := baseUrl + "gists/" + gistId
	if accessToken != "" {
		postUrl = postUrl + "?access_token=" + accessToken
	}

	req, err := http.NewRequest("PATCH", postUrl, jsonBody)
	// handle err
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var jsonErrorMessage GistJSON.MessageResponse
	err = json.Unmarshal(body, &jsonErrorMessage)
	if err != nil {
		return
	}
	if jsonErrorMessage.Message != "" {
		err = errors.New(jsonErrorMessage.Message)
		return
	}

	var jsonRes GistJSON.Response
	err = json.Unmarshal(body, &jsonRes)
	if err != nil {
		return
	}

	fmt.Printf("%s\n", jsonRes.HtmlUrl)
	revisionCount := len(jsonRes.History)
	lastHistoryStatus := jsonRes.History[0].ChangeStatus
	fmt.Printf("Revision %d (%d additions & %d deletions)\n",
		revisionCount, lastHistoryStatus.Deletions, lastHistoryStatus.Additions)

	return
}

func getGistId(urlOrId string) string {
	/*
		accepted gist format are full url: https://gist.github.com/a2a510376da5ffcb93f9
		or just id a2a510376da5ffcb93f9
		split on '/' to retreive only id
	*/
	splitted := strings.Split(urlOrId, "/")
	return splitted[len(splitted)-1]
}
