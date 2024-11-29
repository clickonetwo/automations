/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"context"
	"encoding/gob"
	"os"

	"filippo.io/age"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	SmsHistoryFilename = "sms-history.gob.age"
)

func UploadSmsHistory(events []SmsEvent) error {
	myself, err := age.ParseX25519Recipient(storage.GetConfig().AgePublicKey)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp("", "gob*.age")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	encryptedWriter, err := age.Encrypt(f, myself)
	if err != nil {
		return err
	}
	encodedWriter := gob.NewEncoder(encryptedWriter)
	err = encodedWriter.Encode(events)
	encryptedWriter.Close()
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	err = storage.S3PutBlob(context.Background(), SmsHistoryFilename, f)
	return err
}

func DownloadSmsHistory() ([]SmsEvent, error) {
	f, err := os.CreateTemp("", "gob-*.age")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	err = storage.S3GetBlob(context.Background(), SmsHistoryFilename, f)
	if err != nil {
		return nil, err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	myself, err := age.ParseX25519Identity(storage.GetConfig().AgeSecretKey)
	if err != nil {
		return nil, err
	}
	gobStream, err := age.Decrypt(f, myself)
	if err != nil {
		return nil, err
	}
	var events []SmsEvent
	dec := gob.NewDecoder(gobStream)
	err = dec.Decode(&events)
	if err != nil {
		return nil, err
	}
	return events, nil
}
