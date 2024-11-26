/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"os"
	"testing"

	"github.com/go-test/deep"
)

func TestLoadUsers(t *testing.T) {
	idsEmails, err := ImportUserIdsEmails("../../local/Oasis-admins.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(idsEmails) != 2 {
		t.Errorf("got %d admins, want 2", len(idsEmails))
	}
	if err := Admins.SaveIdsEmails(idsEmails); err != nil {
		t.Fatal(err)
	}
	cpy, err := Admins.IdsEmails()
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(cpy, idsEmails); diff != nil {
		t.Error(diff)
	}
	idsEmails, err = ImportUserIdsEmails("../../local/Oasis-readers.csv")
	if err != nil {
		t.Fatal(err)
	}
	if err := Readers.SaveIdsEmails(idsEmails); err != nil {
		t.Fatal(err)
	}
	t.Logf("Loaded both admins and readers to database")
}

func TestLoginForm(t *testing.T) {
	page := LoginForm("This is a test message.")
	err := os.WriteFile("../../local/test-login-1.html", []byte(page), 0644)
	if err != nil {
		t.Fatal(err)
	}
	page = LoginForm("")
	err = os.WriteFile("../../local/test-login-2.html", []byte(page), 0644)
	if err != nil {
		t.Fatal(err)
	}
}
