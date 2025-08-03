/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/clickonetwo/automations/dialpad/internal/contacts"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

// historyContactsCmd represents the contacts command
var historyContactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "Manage the contacts known to the history server",
	Long: `The history server loads a list of contacts from AWS each time it starts.
This command allows managing that list of contacts.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		envName, _ := cmd.InheritedFlags().GetString("env")
		if err := storage.PushConfig(envName); err != nil {
			panic(err)
		}
		defer storage.PopConfig()
		update, _ := cmd.Flags().GetCount("update")
		export, _ := cmd.Flags().GetString("export")
		var entries []contacts.Entry
		if update > 0 {
			entries = updateContacts()
		}
		if export != "" {
			exportContacts(entries, export)
		}
	},
}

func init() {
	historyCmd.AddCommand(historyContactsCmd)
	historyContactsCmd.Args = cobra.NoArgs
	historyContactsCmd.Flags().Count("update", "update contacts from Dialpad")
	historyContactsCmd.Flags().String("export", "", "export contacts to the specified path")
	historyContactsCmd.MarkFlagsOneRequired("update", "export")
}

func updateContacts() []contacts.Entry {
	log.Printf("Fetching contacts from Dialpad...")
	allContacts, errs := contacts.ListContacts("")
	if errs != nil {
		log.Fatalf("Dialpad fetch failed: %v", errs)
	}
	log.Printf("Uploading contacts to AWS...")
	err := contacts.UploadAllContacts(allContacts)
	if err != nil {
		log.Fatalf("AWS upload failed: %v", err)
	}
	log.Printf("Contacts update complete.")
	return allContacts
}

func exportContacts(entries []contacts.Entry, path string) {
	if entries == nil {
		log.Printf("Fetching users from AWS...")
		dl, err := contacts.DownloadAllContacts()
		if err != nil {
			log.Fatalf("AWS download failed: %v", err)
		}
		entries = dl
	}
	log.Printf("Exporting %d contacts in CSV format...", len(entries))
	if err := contacts.ExportContacts(entries, path); err != nil {
		log.Fatalf("Export failed: %v", err)
	}
	log.Printf("Export of users to %q complete.", path)
}
