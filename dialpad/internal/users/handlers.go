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
	if email, err := CheckAuth(userId, "reader"); err == nil {
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
	if e, err := CheckAuth(userId, "reader"); err == nil && strings.ToLower(e) == strings.ToLower(email) {
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

func DownloadUsers(c *gin.Context) {
	userId := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "Provide authorization"})
		return
	}
	if _, err := CheckAuth(userId, "admin"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "error": err.Error()})
		return
	}
	userType := c.Param("type")
	c.JSON(http.StatusOK, ListUsers(userType))
}

func UploadUsers(c *gin.Context) {
	userId := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "Provide authorization"})
		return
	}
	if _, err := CheckAuth(userId, "admin"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "error": err.Error()})
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	var userMap map[string]string
	if err = json.Unmarshal(body, &userMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}
	userType := c.Param("type")
	switch userType {
	case "admin":
		err = Admins.SaveIdsEmails(userMap)
	case "reader":
		err = Readers.SaveIdsEmails(userMap)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("unknown user type: %s", userType)})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "created", "type": userType, "count": len(userMap)})
}
