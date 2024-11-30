/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"context"
	"encoding/gob"
	"os"

	"filippo.io/age"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	usersFilename = "users.gob.age"
)

func UploadUsersList(entries []Entry) error {
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
	err = storage.S3PutBlob(context.Background(), usersFilename, f)
	return err
}

func DownloadUsersList() ([]Entry, error) {
	f, err := os.CreateTemp("", "gob-*.age")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	err = storage.S3GetBlob(context.Background(), usersFilename, f)
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
	var events []Entry
	dec := gob.NewDecoder(gobStream)
	err = dec.Decode(&events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// LoadUsers loads the users list from AWS.
//
// SuperAdmins in Dialpad become admins in this server.
// Every user becomes a reader in this server.
func LoadUsers() error {
	users, err := DownloadUsersList()
	if err != nil {
		return err
	}
	Readers, Admins = make(map[string]Entry), make(map[string]Entry)
	for _, user := range users {
		Readers[user.Id] = user
		if user.IsSuperAdmin {
			Admins[user.Id] = user
		}
	}
	return nil
}
