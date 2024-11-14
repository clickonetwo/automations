/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package event

import (
	"testing"

	"github.com/clickonetwo/automations/dialpad/internal/auth"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

func TestCreateFindDeleteHook(t *testing.T) {
	// we need a working API key, so use dev environment to test
	if err := storage.PushConfig("development"); err != nil {
		t.Fatal(err)
	}
	c, _ := middleware.CreateTestContext()
	hookUrl := "https://dev0.clickonetwo.io/path/to/hook"
	secret := auth.MakeNonce()
	id, err := FindHook(c, hookUrl, secret)
	if err != nil {
		t.Fatal(err)
	}
	if id != "" {
		t.Fatalf("missing hook id should be non-empty, got %s", id)
	}
	id, err = CreateHook(c, hookUrl, secret)
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatalf("created hook id should be non-empty")
	}
	id, err = FindHook(c, hookUrl, secret)
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatalf("found hook id should be non-empty")
	}
	err = DeleteHook(c, id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnsureWebhook(t *testing.T) {
	// we need a working API key, so use dev environment to test
	if err := storage.PushConfig("development"); err != nil {
		t.Fatal(err)
	}
	c, _ := middleware.CreateTestContext()
	id, err := EnsureWebHook(c, "/hook/me", "fake secret")
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatalf("hook id should be non-empty")
	}
	id2, err := EnsureWebHook(c, "/hook/me", "fake secret")
	if err != nil {
		t.Fatal(err)
	}
	if id != id2 {
		t.Errorf("hook id should be %s, got %s", id, id2)
		err = DeleteHook(c, id2)
		if err != nil {
			t.Error(err)
		}
	}
	err = DeleteHook(c, id)
	if err != nil {
		t.Fatal(err)
	}
}
