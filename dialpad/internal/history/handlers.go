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
	"time"

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
	email := users.CheckAuth(userId, "reader")
	if email == "" {
		c.Data(http.StatusOK, "text/html", ServerErrorForm(name, phone))
		return
	}
	thread := SelectThreadByEmailPhone(email, phone, EventHistory)
	c.Data(http.StatusOK, "text/html", RequestForm(name, phone, thread))
}

func SearchHandler(c *gin.Context) {
	filter := c.Query("filter")
	userId, _ := c.Cookie(users.AuthCookieName)
	email := users.CheckAuth(userId, "reader")
	if email == "" {
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

func StatsHandler(c *gin.Context) {
	userId, _ := c.Cookie(users.AuthCookieName)
	if userId == "" {
		c.Redirect(http.StatusFound, "/login?next=stats")
		return
	}
	if email := users.CheckAuth(userId, "admin"); email != "" {
		stats, err := users.UsageStats.FetchAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "details": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{
			"event_count":   len(EventHistory),
			"first_event":   time.UnixMicro(EventHistory[0].Date).In(PT).Format(time.RFC1123),
			"last_event":    time.UnixMicro(EventHistory[len(EventHistory)-1].Date).In(PT).Format(time.RFC1123),
			"contact_count": len(AllContacts),
			"reader_count":  len(users.ListUsers("reader")),
			"admin_count":   len(users.ListUsers("admin")),
			"access_counts": stats,
		})
	} else {
		c.Redirect(http.StatusFound, "/login?next=stats")
	}
}
