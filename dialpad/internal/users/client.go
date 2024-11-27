/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/clickonetwo/automations/dialpad/internal/contacts"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

type Entry struct {
	Id           string   `json:"id"`
	Emails       []string `json:"emails"`
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	IsAdmin      bool     `json:"is_admin"`
	IsSuperAdmin bool     `json:"is_super_admin"`
}

type entryPage struct {
	Cursor string  `json:"cursor"`
	Items  []Entry `json:"items"`
}

func DownloadUsers() ([]Entry, error) {
	var results []Entry
	key := storage.GetConfig().DialpadApiKey
	baseUrl := fmt.Sprintf("%s/users?limit=200&apikey=%s", contacts.DialpadApiRoot, key)
	cursor := ""
	for {
		url := baseUrl
		if cursor != "" {
			url = fmt.Sprintf("%s&cursor=%s", url, cursor)
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("accept", "application/json")
		resp, err := contacts.DialPadListClient.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("status code: %d, body: %s", resp.StatusCode, body)
		}
		var result entryPage
		err = json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}
		if len(result.Items) == 0 {
			break
		}
		for _, entry := range result.Items {
			results = append(results, entry)
		}
		cursor = result.Cursor
		if cursor == "" {
			break
		}
	}
	return results, nil
}

// LoadUsers gets the latest users from Dialpad.
//
// SuperAdmins in Dialpad become admins in this server.
// Every user becomes a reader in this server.
func LoadUsers() error {
	users, err := DownloadUsers()
	if err != nil {
		return err
	}
	Readers, Admins = make(map[string]Entry), make(map[string]Entry)
	for _, user := range users {
		Readers[user.Id] = user
		if user.IsSuperAdmin {
			Admins[user.Id] = user
		}
	}
	return nil
}
