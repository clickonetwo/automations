/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/clickonetwo/automations/dialpad/internal/contacts"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
	"github.com/clickonetwo/automations/dialpad/internal/users"
)

var (
	EventHistory []SmsEvent
	AllContacts  []contacts.Entry
)

func RequestHandler(c *gin.Context) {
	phone := c.Query("phone")
	name := c.Query("name")
	if phone == "" {
		c.Redirect(http.StatusFound, fmt.Sprintf("/search?filter=%s", url.QueryEscape(name)))
		return
	}
	if name == "" {
		name = contacts.UnknownName
	}
	userId, _ := c.Cookie(users.AuthCookieName)
	email, err := users.CheckAuth(userId, "reader")
	if err != nil {
		c.Data(http.StatusOK, "text/html", ServerErrorForm(name, phone))
		return
	}
	thread := SelectThreadByEmailPhone(email, phone, EventHistory)
	c.Data(http.StatusOK, "text/html", RequestForm(name, phone, thread))
}

func SearchHandler(c *gin.Context) {
	filter := c.Query("filter")
	userId, _ := c.Cookie(users.AuthCookieName)
	email, err := users.CheckAuth(userId, "reader")
	if err != nil {
		c.Data(http.StatusOK, "text/html", contacts.ServerErrorForm(filter))
		return
	}
	phones := SelectPhonesByEmail(email, EventHistory)
	entries := contacts.SelectEntriesByPhones(phones, AllContacts)
	if filter == "" {
		c.Data(http.StatusOK, "text/html", contacts.SearchForm("", entries))
		return
	}
	entries = contacts.FilterSearchEntries(filter, entries)
	c.Data(http.StatusOK, "text/html", contacts.SearchForm(filter, entries))
}

func LoadEventHistory() error {
	dir, err := storage.FindEnvFile("data", true)
	if err != nil {
		return err
	}
	events, err := ImportEncryptedSmsEvents(dir + "data/all-events.csv.age")
	if err != nil {
		return err
	}
	EventHistory = events
	return nil
}

func LoadAllContacts() error {
	dir, err := storage.FindEnvFile("data", true)
	if err != nil {
		return err
	}
	entries, err := contacts.ImportEncryptedContacts(dir + "data/all-contacts.csv.age")
	if err != nil {
		return err
	}
	AllContacts = entries
	return nil
}
