/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"testing"
)

func TestImportSmsEvents(t *testing.T) {
	result, err := ImportSmsEvents("../../local/texts-all.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 158113 {
		t.Errorf("Expected %d rows, got %d", 0, len(result))
	}
}

func TestSelectThreadMatchingEmailAndPhone(t *testing.T) {
	events, err := ImportSmsEvents("../../local/texts-all.csv")
	if err != nil {
		t.Fatal(err)
	}
	thread := SelectThreadByEmailPhone("anuar.arriaga@oasislegalservices.org", "+14158234525", events)
	if len(thread) != 28 {
		t.Errorf("Expected %d items in thread, got %d", 0, len(thread))
	}
}
