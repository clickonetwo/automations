/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/go-test/deep"
	"github.com/schollz/progressbar/v3"

	"github.com/clickonetwo/automations/dialpad/internal/contacts"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	SmsImportHeaders = []string{
		"date", "message_id", "name", "email", "target_type", "target_id",
		"sender_id", "direction", "to_phone", "from_phone",
		"encrypted_text", "encrypted_aes_text", "mms", "timezone",
	}
)

func ImportSmsEvents(path string) ([]SmsEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := contacts.BOMAwareCSVReader(f)
	record, err := reader.Read()
	if diff := deep.Equal(record, SmsImportHeaders); diff != nil {
		return nil, fmt.Errorf("unexpected column names: %v", record)
	}
	return parseSmsEvents(reader)
}

func ImportEncryptedSmsEvents(path string) ([]SmsEvent, error) {
	id, err := age.ParseX25519Identity(storage.GetConfig().AgeSecretKey)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	df, err := age.Decrypt(f, id)
	if err != nil {
		return nil, err
	}
	reader := contacts.BOMAwareCSVReader(df)
	record, err := reader.Read()
	if diff := deep.Equal(record, SmsImportHeaders); diff != nil {
		return nil, fmt.Errorf("unexpected column names: %v", record)
	}
	return parseSmsEvents(reader)
}

func LoadSmsEvents() error {
	dir, err := storage.FindEnvFile("data", true)
	if err != nil {
		return err
	}
	events, err := ImportEncryptedSmsEvents(dir + "data/all-events.csv.age")
	if err != nil {
		return err
	}
	EventHistory = events
	return nil
}

func parseSmsEvents(reader *csv.Reader) ([]SmsEvent, error) {
	var events []SmsEvent
	var event SmsEvent
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
			log.Panicf("error reading record from csv: %v", err)
		}
		if len(record) != len(SmsImportHeaders) {
			return nil, fmt.Errorf("row %d has %d columns, expected %d", row, len(record), len(SmsImportHeaders))
		}
		if record[13] != "UTC" {
			return nil, fmt.Errorf("row %d has an invalid timezone: %q", row, record[13])
		}
		date, err := time.Parse("2006-01-02 15:04:05", record[0])
		if err != nil {
			return nil, fmt.Errorf("row %d has an invalid date (%s): %v", row, record[0], err)
		}
		event.Date = date.UnixMicro()
		event.MessageId = record[1]
		event.Name = record[2]
		event.Email = record[3]
		event.TargetType = record[4]
		event.TargetId, _ = strconv.ParseInt(record[5], 10, 64)
		event.SenderId, _ = strconv.ParseInt(record[6], 10, 64)
		event.Direction = record[7]
		event.ToPhones = strings.Split(record[8], ",")
		event.FromPhone = record[9]
		if record[10] != "" {
			return nil, fmt.Errorf("row %d has a non-empty encrypted text: %q", row, record[10])
		}
		event.Text = record[11]
		event.MmsUrl = record[12]
		events = append(events, event)
	}
	return events, nil
}
