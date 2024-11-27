/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	AuthCookieName = "dialpad_history_user_id"
	Admins         map[string]Entry
	Readers        map[string]Entry
)

func ListUsers(capability string) map[string]Entry {
	switch capability {
	case "admin":
		return Admins
	case "reader":
		return Readers
	default:
		return nil
	}
}

func CheckAuth(userId, capability string) string {
	env := storage.GetConfig()
	if userId == env.MasterAdminId {
		return env.MasterAdminEmail
	}
	switch capability {
	case "admin":
		if admin, ok := Admins[userId]; ok {
			return admin.Emails[0]
		}
	case "reader":
		if reader, ok := Readers[userId]; ok {
			return reader.Emails[0]
		}
	}
	return ""
}
