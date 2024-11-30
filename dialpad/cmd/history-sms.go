/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"log"
	"strings"

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
		if initialize > 0 {
			initializeHistory(args[0])
		} else if update > 0 {
			updateHistory(args[0])
		} else {
			log.Fatalf("You must specify one of --update or --initialize and a report path to import")
		}
	},
}

func init() {
	historyCmd.AddCommand(smsCmd)
	smsCmd.Args = cobra.ExactArgs(1)
	smsCmd.Flags().Count("initialize", "initialize SMS history from report")
	smsCmd.Flags().Count("update", "update SMS history from report")
	smsCmd.MarkFlagsOneRequired("initialize", "update")
	smsCmd.MarkFlagsMutuallyExclusive("initialize", "update")
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
