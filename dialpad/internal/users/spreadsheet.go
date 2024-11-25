/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/schollz/progressbar/v3"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

func ImportUserIdsEmails(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := storage.BOMAwareCSVReader(f)
	record, err := reader.Read()
	if record[0] != "target_id" || record[1] != "type" || record[2] != "primary_email" {
		return nil, fmt.Errorf("unexpected column names: %v", record)
	}
	return parseIdsEmails(reader)
}

func ExportUserIdsEmails(path string, userMap map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	err = w.Write([]string{"target_id", "type", "primary_email"})
	if err != nil {
		return err
	}
	for k, v := range userMap {
		err = w.Write([]string{k, "user", v})
		if err != nil {
			return err
		}
	}
	return nil
}

func parseIdsEmails(reader *csv.Reader) (map[string]string, error) {
	results := make(map[string]string)
	var row = 1
	bar := progressbar.Default(-1, "Validating entries")
	defer bar.Close()
	for {
		row++
		_ = bar.Add(1)
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("error reading record from csv: %v", err)
		}
		if len(record) < 3 {
			return nil, fmt.Errorf("row %d has %d columns, expected at least %d", row, len(record), 3)
		}
		if record[1] != "user" {
			continue
		}
		results[record[0]] = record[2]
	}
	return results, nil
}
