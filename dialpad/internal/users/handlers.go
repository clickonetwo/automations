/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	dataMissingMsg = "You must specify both username and password"
	badDataMsg     = "Invalid credentials. Please try again."
	loginAge       = 365 * 24 * 60 * 60 // 1 year
	UsageStats     = middleware.StatMap("usage-status")
)

func CheckLoginMiddleware(c *gin.Context) {
	userId, _ := c.Cookie(AuthCookieName)
	if userId == "" {
		noAuth, _ := UsageStats.Int64("unauthenticated requests")
		_ = UsageStats.SetInt64("unauthenticated requests", noAuth+1)
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if email := CheckAuth(userId, "reader"); email != "" {
		auth, err := UsageStats.MapInt64("authenticated requests")
		if err == nil {
			auth[email] += 1
			_ = UsageStats.SetMapInt64("authenticated requests", auth)
		}
		c.Next()
	} else {
		noAuth, _ := UsageStats.Int64("unauthenticated requests")
		_ = UsageStats.SetInt64("unauthenticated requests", noAuth+1)
		c.Redirect(http.StatusFound, "/login")
	}
}

func LoginHandler(c *gin.Context) {
	secure := storage.GetConfig().Name != "development"
	email := c.Query("username")
	userId := c.Query("password")
	next := c.Query("next")
	if email == "" && userId == "" {
		var message string
		if userId, _ := c.Cookie(AuthCookieName); userId != "" {
			c.SetCookie(AuthCookieName, "", -1, "/", "", secure, true)
			message = "Please log in again."
		}
		c.Data(http.StatusOK, "text/html", LoginForm(message, next))
		return
	}
	if email == "" || userId == "" {
		c.SetCookie(AuthCookieName, "", -1, "/", "", secure, true)
		c.Data(http.StatusOK, "text/html", LoginForm(dataMissingMsg, next))
		return
	}
	if e := CheckAuth(userId, "reader"); e != "" && strings.ToLower(e) == strings.ToLower(email) {
		c.SetCookie(AuthCookieName, userId, loginAge, "/", "", secure, true)
		c.Data(http.StatusOK, "text/html", LoginSuccessForm(next))
	} else {
		c.SetCookie(AuthCookieName, "", -1, "/", "", secure, true)
		c.Data(http.StatusOK, "text/html", LoginForm(badDataMsg, next))
	}
}

func LogoutHandler(c *gin.Context) {
	secure := storage.GetConfig().Name != "development"
	c.SetCookie(AuthCookieName, "", -1, "/", "", secure, true)
	c.Redirect(http.StatusFound, "/login")
}
