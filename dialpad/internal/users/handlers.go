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

	"github.com/gin-gonic/gin"
)

func UploadUsers(c *gin.Context) {
	userId := c.Query("user")
	if _, err := CheckAuth(userId, "admin"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "error": err.Error()})
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
