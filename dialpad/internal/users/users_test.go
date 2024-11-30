/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"os"
	"testing"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

func TestLoginForm(t *testing.T) {
	page := LoginForm("This is a test message.", "")
	err := os.WriteFile("../../local/test-login-1.html", []byte(page), 0644)
	if err != nil {
		t.Fatal(err)
	}
	page = LoginForm("", "test")
	err = os.WriteFile("../../local/test-login-2.html", []byte(page), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadUsers(t *testing.T) {
	err := storage.PushConfig("development")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.PopConfig()
	if err := LoadUsers(); err != nil {
		t.Fatal(err)
	}
}
