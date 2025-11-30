/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package migration

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCleanupPhones(t *testing.T) {
	phones, err := ExtractOneFieldFromFile("../../local/all-inquiries-table.csv", "Phone Number")
	if err != nil {
		t.Fatal(err)
	}
	clean := make(map[string]string)
	dirty := make([]string, 0, 20)
	for _, phone := range phones {
		valid := findValidPhone(phone)
		if valid == "" {
			if strings.TrimSpace(phone) != "" {
				dirty = append(dirty, phone)
			}
		} else {
			clean[phone] = valid
		}
	}
	exportJsonToPath(clean, "/tmp/clean-phones.json")
	if len(dirty) > 0 {
		t.Errorf("got %d dirty phones", len(dirty))
		exportJsonToPath(dirty, "/tmp/dirty-phones.json")
	}
}

func TestCleanupStates(t *testing.T) {
	states, err := ExtractOneFieldFromFile("../../local/all-inquiries-table.csv", "State / Province")
	if err != nil {
		t.Fatal(err)
	}
	clean := make(map[string]string)
	dirty := make([]string, 0, 20)
	for _, state := range states {
		valid := findValidState(state)
		if valid == "" {
			if strings.TrimSpace(state) != "" {
				dirty = append(dirty, state)
			}
		} else {
			clean[state] = valid
		}
	}
	exportJsonToPath(clean, "/tmp/clean-states.json")
	if len(dirty) > 0 {
		t.Errorf("got %d dirty states", len(dirty))
		exportJsonToPath(dirty, "/tmp/dirty-states.json")
	}
}

func exportJsonToPath(obj any, path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	err = e.Encode(obj)
	if err != nil {
		panic(err)
	}
}
