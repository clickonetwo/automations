/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package contacts

import (
	"context"
	"encoding/gob"
	"os"

	"filippo.io/age"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	AllContactsFilename = "contacts.gob.age"
)

func UploadAllContacts(entries []Entry) error {
	myself, err := age.ParseX25519Recipient(storage.GetConfig().AgePublicKey)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp("", "gob-*.age")
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
	err = encodedWriter.Encode(entries)
	encryptedWriter.Close()
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	err = storage.S3PutBlob(context.Background(), AllContactsFilename, f)
	return err
}

func DownloadAllContacts() ([]Entry, error) {
	f, err := os.CreateTemp("", "gob-*.age")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	err = storage.S3GetBlob(context.Background(), AllContactsFilename, f)
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
	var entries []Entry
	dec := gob.NewDecoder(gobStream)
	err = dec.Decode(&entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}
