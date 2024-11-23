/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"os"
	"testing"
)

func TestFormatThread(t *testing.T) {
	events, err := ImportSmsEvents("../../local/texts-all.csv")
	if err != nil {
		t.Fatal(err)
	}
	thread := SelectThreadByEmailPhone("anuar.arriaga@oasislegalservices.org", "+14158234525", events)
	page := FormatThread("+14158234525", thread)
	err = os.WriteFile("../../local/test-thread.html", []byte(page), 0644)
	if err != nil {
		t.Fatal(err)
	}
}
