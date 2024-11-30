/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"io"
	"net/http"
	"os"

	"filippo.io/age"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

func DownloadAndEncryptSmsReport(url, path string) error {
	recipient, err := age.ParseX25519Recipient(storage.GetConfig().AgePublicKey)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w, err := age.Encrypt(f, recipient)
	if err != nil {
		return err
	}
	defer w.Close()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
