/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/clickonetwo/automations/dialpad/internal/users"
)

var (
	EventHistory []SmsEvent
)

func RequestHandler(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.String(http.StatusOK, "%s", RequestForm("", nil))
		return
	}
	userId, _ := c.Cookie(users.AuthCookieName)
	email, err := users.CheckAuth(userId, "reader")
	if err != nil {
		c.String(http.StatusOK, "%s", ServerErrorForm(phone))
		return
	}
	thread := SelectThreadByEmailPhone(email, phone, EventHistory)
	c.String(http.StatusOK, "%s", RequestForm(phone, thread))
}
