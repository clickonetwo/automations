/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/clickonetwo/automations/dialpad/internal/history"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

// smsCmd represents the sms command
var smsCmd = &cobra.Command{
	Use:   "sms [flags] path-to-report.csv[.age]",
	Short: "Manage the SMS event history known to the history server",
	Long: `The history server loads an SMS event history from AWS each time it starts.
This command allows initializing or updating that history from a downloaded SMS report.
If the report path ends in ".age", it's decrypted before being imported.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		envName, _ := cmd.InheritedFlags().GetString("env")
		if err := storage.PushConfig(envName); err != nil {
			panic(err)
		}
		defer storage.PopConfig()
		update, _ := cmd.Flags().GetCount("update")
		initialize, _ := cmd.Flags().GetCount("initialize")
		id, _ := cmd.Flags().GetString("download-report")
		if initialize > 0 {
			initializeHistory(args[0])
		} else if update > 0 {
			updateHistory(args[0])
		} else {
			requestReport(id, args[0])
		}
	},
}

func init() {
	historyCmd.AddCommand(smsCmd)
	smsCmd.Args = cobra.ExactArgs(1)
	smsCmd.Flags().Count("initialize", "initialize SMS history from report")
	smsCmd.Flags().Count("update", "update SMS history from report")
	smsCmd.Flags().Count("request-report", "request report for SMS history update")
	smsCmd.Flags().String("download-report", "", "download requested SMS history report")
	smsCmd.MarkFlagsOneRequired("initialize", "update", "request-report", "download-report")
	smsCmd.MarkFlagsMutuallyExclusive("initialize", "update", "request-report", "download-report")
}

func initializeHistory(path string) {
	log.Printf("Importing SMS history from %q...", path)
	var events []history.SmsEvent
	var err error
	if strings.HasSuffix(path, ".age") {
		events, err = history.ImportEncryptedSmsEvents(path)
	} else {
		events, err = history.ImportSmsEvents(path)
	}
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}
	log.Printf("Uploading SMS history (%d events) to AWS...", len(events))
	err = history.UploadSmsHistory(events)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	log.Printf("Successfully initialized the saved SMS history.")
}

func updateHistory(path string) {
	log.Printf("Downloading existing SMS history...")
	err := history.LoadEventHistory()
	if err != nil {
		log.Fatalf("Failed to load existing history: %v", err)
	}
	log.Printf("Importing additional history from %q...", path)
	var events []history.SmsEvent
	if strings.HasSuffix(path, ".age") {
		events, err = history.ImportEncryptedSmsEvents(path)
	} else {
		events, err = history.ImportSmsEvents(path)
	}
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}
	log.Printf("Successfully imported %d SMS events, starting merge...", len(events))
	earliest := history.EventHistory[len(history.EventHistory)-1].Date
	firstLater := -1
	for index, event := range events {
		if event.Date > earliest {
			firstLater = index
			break
		}
	}
	if firstLater == -1 {
		log.Fatalf("No events later than existing found to import")
	}
	log.Printf("Merging %d later events and uploading to AWS...", len(events)-firstLater)
	history.EventHistory = append(history.EventHistory, events[firstLater:]...)
	err = history.UploadSmsHistory(events)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	log.Printf("Successfully uploaded the updated SMS history.")
}

func requestReport(id, path string) {
	if id == "" {
		log.Printf("Downloading existing SMS history...")
		err := history.LoadEventHistory()
		if err != nil {
			log.Fatalf("Failed to load existing history: %v", err)
		}
		earliest := history.EventHistory[len(history.EventHistory)-1].Date
		log.Printf("Last SMS in history is dated %s", time.UnixMicro(earliest).Format(time.RFC1123))
		log.Printf("Requesting a report for all days since then...")
		id, err = history.RequestSmsReport(earliest)
		if err != nil {
			log.Fatalf("Request failed: %v", err)
		}
	}
	log.Printf("Waiting for report with id %s (^C to stop)...", id)
	var url string
	var err error
	bar := progressbar.Default(-1, "Waiting for report")
	for wait := false; url == "" && err == nil; wait = true {
		if wait {
			for i := 0; i < 20; i++ {
				time.Sleep(time.Second / 4)
				_ = bar.Add(1)
			}
		}
		url, err = history.GetSmsReportDownloadUrl(id)
	}
	_ = bar.Close()
	if err != nil {
		log.Fatalf("Couldn't wait for report, try again with --download-report %s: %v", id, err)
	}
	log.Printf("Downloading report with id %s...", id)
	err = history.DownloadSmsReport(url, path, false)
	if err != nil {
		log.Fatalf("Download failed, try again with --download-report %s: %v", id, err)
	}
	log.Printf("Successfully downloaded report with id %s to %q", id, path)
}
