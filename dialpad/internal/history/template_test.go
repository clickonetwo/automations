/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"os"
	"testing"

	"github.com/clickonetwo/automations/dialpad/internal/contacts"
)

func TestHistoryForm(t *testing.T) {
	events, err := ImportSmsEvents("../../local/texts-all.csv")
	if err != nil {
		t.Fatal(err)
	}
	thread := SelectThreadByEmailPhone("anuar.arriaga@oasislegalservices.org", "+14158234525", events)
	page := RequestForm("Moises Someone", "+14158234525", thread)
	err = os.WriteFile("../../local/test-thread-1.html", []byte(page), 0644)
	if err != nil {
		t.Fatal(err)
	}
	thread = SelectThreadByEmailPhone("anuar.arriaga@oasislegalservices.org", "+15109260499", events)
	page = RequestForm("Daniel Brotsky", "+15109260499", thread)
	err = os.WriteFile("../../local/test-thread-2.html", []byte(page), 0644)
	if err != nil {
		t.Fatal(err)
	}
	page = RequestForm(contacts.UnknownName, "+15109260499", nil)
	err = os.WriteFile("../../local/test-thread-3.html", []byte(page), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHistoryServerError(t *testing.T) {
	page := ServerErrorForm(`Daniel Brotsky`, `+15109260499`)
	err := os.WriteFile("../../local/test-error-1.html", page, 0644)
	if err != nil {
		t.Fatal(err)
	}
	page = ServerErrorForm(`{unknown}`, `+15109260499`)
	err = os.WriteFile("../../local/test-error-2.html", page, 0644)
	if err != nil {
		t.Fatal(err)
	}
	page = ServerErrorForm(``, `+15109260499`)
	err = os.WriteFile("../../local/test-error-2.html", page, 0644)
	if err != nil {
		t.Fatal(err)
	}
}
