/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"filippo.io/age"

	"github.com/clickonetwo/automations/dialpad/internal/contacts"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

func DownloadSmsReport(url, path string, encrypt bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := io.WriteCloser(f)
	if encrypt {
		recipient, err := age.ParseX25519Recipient(storage.GetConfig().AgePublicKey)
		if err != nil {
			return err
		}
		w, err = age.Encrypt(f, recipient)
		if err != nil {
			return err
		}
		defer w.Close()
	}
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

type reportRequest struct {
	DaysAgoEarliest int64  `json:"days_ago_start"`
	DaysAgoLatest   int64  `json:"days_ago_end"`
	ExportType      string `json:"export_type"`
	StatType        string `json:"stat_type"`
	Timezone        string `json:"timezone"`
}

type reportResponse struct {
	AlreadyStarted bool   `json:"already_started"`
	RequestId      string `json:"request_id"`
}

func RequestSmsReport(fromDate int64) (string, error) {
	var usecPerDay int64 = 1_000_000 * 60 * 60 * 24
	usec := time.Now().UnixMicro() - fromDate
	if usec < 2*usecPerDay {
		return "", errors.New("can't report on less than two days")
	}
	apiUrl := fmt.Sprintf("%s/stats?apikey=%s", contacts.DialpadApiRoot, storage.GetConfig().DialpadApiKey)
	body, err := json.Marshal(reportRequest{
		DaysAgoEarliest: (usec / usecPerDay) + 1,
		DaysAgoLatest:   1,
		ExportType:      "records",
		StatType:        "texts",
		Timezone:        "UTC",
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, apiUrl, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("report request failed: status %s, %s", resp.Status, body)
	}
	var result reportResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("report response not understood: %v", err)
	}
	return result.RequestId, nil
}

type downloadResponse struct {
	DownloadUrl string `json:"download_url"`
	FileType    string `json:"file_type"`
	Status      string `json:"status"`
}

func GetSmsReportDownloadUrl(id string) (string, error) {
	apiUrl := fmt.Sprintf("%s/stats/%s?apikey=%s", contacts.DialpadApiRoot, id, storage.GetConfig().DialpadApiKey)
	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download url request failed: status %s, %s", resp.Status, body)
	}
	var result downloadResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("download response not understood: %v", err)
	}
	if result.Status != "complete" {
		return "", nil
	}
	return result.DownloadUrl, nil
}
