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

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	dataMissingMsg = "You must specify both username and password"
	badDataMsg     = "Invalid credentials. Please try again."
	loginAge       = 365 * 24 * 60 * 60 // 1 year
)

func CheckLoginMiddleware(c *gin.Context) {
	userId, _ := c.Cookie(AuthCookieName)
	if userId == "" {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if _, err := CheckAuth(userId, "reader"); err == nil {
		c.Next()
	} else {
		c.Redirect(http.StatusFound, "/login")
	}
}

func LoginHandler(c *gin.Context) {
	secure := storage.GetConfig().Name != "development"
	email := c.Query("username")
	userId := c.Query("password")
	if email == "" && userId == "" {
		var message string
		if userId, _ := c.Cookie(AuthCookieName); userId != "" {
			c.SetCookie(AuthCookieName, "", -1, "/", "", secure, true)
			message = "Please log in again."
		}
		c.Data(http.StatusOK, "text/html", LoginForm(message))
		return
	}
	if email == "" || userId == "" {
		c.SetCookie(AuthCookieName, "", -1, "/", "", secure, true)
		c.Data(http.StatusOK, "text/html", LoginForm(dataMissingMsg))
		return
	}
	if e, err := CheckAuth(userId, "reader"); err == nil && strings.ToLower(e) == strings.ToLower(email) {
		c.SetCookie(AuthCookieName, userId, loginAge, "/", "", secure, true)
		c.Data(http.StatusOK, "text/html", LoginSuccessForm())
	} else {
		c.SetCookie(AuthCookieName, "", -1, "/", "", secure, true)
		c.Data(http.StatusOK, "text/html", LoginForm(badDataMsg))
	}
}

func LogoutHandler(c *gin.Context) {
	secure := storage.GetConfig().Name != "development"
	c.SetCookie(AuthCookieName, "", -1, "/", "", secure, true)
	c.Redirect(http.StatusFound, "/login")
}

func DownloadUsers(c *gin.Context) {
	userId := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if _, err := CheckAuth(userId, "admin"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "error": err.Error()})
		return
	}
	userType := c.Param("type")
	c.JSON(http.StatusOK, ListUsers(userType))
}

func UploadUsers(c *gin.Context) {
	userId := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
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
