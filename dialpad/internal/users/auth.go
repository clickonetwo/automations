/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"errors"
	"fmt"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	AuthCookieName = "dialpad_history_user_id"
)

func ListUsers(capability string) map[string]string {
	var userMap map[string]string
	switch capability {
	case "admin":
		userMap, _ = Admins.IdsEmails()
	case "reader":
		userMap, _ = Readers.IdsEmails()
	}
	return userMap
}

func CheckAuth(userId, capability string) (string, error) {
	env := storage.GetConfig()
	if userId == env.MasterAdminId {
		return env.MasterAdminEmail, nil
	}
	// not the master user, so check the database
	var userMap map[string]string
	var err error
	switch capability {
	case "admin":
		userMap, err = Admins.IdsEmails()
	case "reader":
		userMap, err = Readers.IdsEmails()
	case "default":
		err = fmt.Errorf("unknown capability: %s", capability)
	}
	if err != nil {
		return "", err
	}
	email, ok := userMap[userId]
	if !ok {
		return "", errors.New("forbidden")
	}
	return email, nil
}
