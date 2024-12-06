/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package migration

import (
	"encoding/csv"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestFromToCopyMap(t *testing.T) {
	toRow := make(map[string]string)
	for _, fromCol := range fromFieldNames {
		if toCol, ok := fromToCopyMap[fromCol]; !ok {
			t.Errorf("From column %q not found in fromToCopyMap", fromCol)
		} else if strings.HasPrefix(toCol, "*") {
			if toCol == "*clean*" {
				if err := cleanField(fromCol, "", toRow); err != nil {
					t.Error(err)
				} else {
					for key := range toRow {
						if !slices.Contains(toFieldNames, key) {
							t.Errorf("Cleaned column %q not found in toFieldNames", key)
						}
					}
				}
			} else if toCol != "*ignore*" {
				t.Errorf("From column %q maps to unknown directive %q", fromCol, toCol)
			}
		} else if !slices.Contains(toFieldNames, toCol) {
			t.Errorf("To column %q not found in toFieldNames", toCol)
		}
	}
}

func TestImportAndClean(t *testing.T) {
	in, err := os.Open("../../local/all-inquiries-table.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	r := BOMAwareCSVReader(in)
	out, err := os.Create("../../local/all-contacts-master-table.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	w := csv.NewWriter(out)
	defer w.Flush()
	err = ImportAndClean(r, w)
	if err != nil {
		t.Fatal(err)
	}
}
