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

	"github.com/clickonetwo/automations/dialpad/internal/contacts"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download [flags] path_to_csv",
	Short: "Download contacts",
	Long: `Downloads all the contacts in Dialpad to a local spreadsheet.
You must specify the path to the CSV-format spreadsheet to be created.
If no account id is specified, downloads the company contacts.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Default().SetFlags(0)
		accountId, err := cmd.Flags().GetString("account")
		if err != nil {
			log.Panic(err)
		}
		download(args[0], accountId)
	},
}

func init() {
	contactsCmd.AddCommand(downloadCmd)

	downloadCmd.Args = cobra.ExactArgs(1)
	downloadCmd.Flags().StringP("account", "a", "", "User account id")
}

func download(path string, accountId string) {
	err := storage.PushConfig("")
	if err != nil {
		log.Fatalf("You must have a .env file containg the Dialpad API key")
	}
	defer storage.PopConfig()
	entries, err := contacts.ListContacts(accountId)
	if err != nil {
		log.Printf("Download was interrupted: %v", err)
		path = strings.TrimSuffix(path, ".csv") + ".partial.csv"
		log.Printf("Saving partial results (%d entries) to %s", len(entries), path)
	} else {
		log.Printf("Saving results (%d entries) to %s", len(entries), path)
	}
	if err := contacts.ExportContacts(entries, path); err != nil {
		log.Fatalf("Could not write file at path %s: %v", path, err)
	}
}
