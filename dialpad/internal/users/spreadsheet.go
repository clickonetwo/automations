/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

func ExportUsers(entries []Entry, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	err = w.Write([]string{"Full Name", "Is Admin?", "Login Link"})
	if err != nil {
		return err
	}
	slices.SortFunc(entries, CompareEntries)
	for _, e := range entries {
		name := e.FirstName + " " + e.LastName
		admin := "No"
		if e.IsSuperAdmin {
			admin = "Yes"
		}
		link := fmt.Sprintf("%s/login?username=%s&password=%s",
			storage.GetConfig().HerokuHostUrl, url.QueryEscape(e.Emails[0]), url.QueryEscape(e.Id),
		)
		err = w.Write([]string{name, admin, link})
		if err != nil {
			return err
		}
	}
	return nil
}

func CompareEntries(e1, e2 Entry) int {
	n1 := e1.FirstName + " " + e2.LastName
	n2 := e2.FirstName + " " + e2.LastName
	return strings.Compare(strings.ToLower(n1), strings.ToLower(n2))
}
