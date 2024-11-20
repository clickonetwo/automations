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

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete contacts",
	Long: `Deletes contacts that match criteria specified by a flag.
See the flags for the details of the criteria.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Default().SetFlags(0)
		drCount, _ := cmd.Flags().GetCount("dry-run")
		offset, _ := cmd.Flags().GetInt64("duplicate-offset")
		wpCount, _ := cmd.Flags().GetCount("without-phones")
		deleteContacts(drCount > 0, offset, wpCount > 0)
	},
}

func init() {
	contactsCmd.AddCommand(deleteCmd)

	deleteCmd.Args = cobra.ExactArgs(0)
	deleteCmd.Flags().CountP("dry-run", "d", "Don't delete, just report what would be deleted")
	deleteCmd.Flags().Int64("duplicate-offset", 0, "Delete duplicates with UIDs offset by some value")
	deleteCmd.Flags().Count("without-phones", "Delete contacts that have no phone numbers")
	deleteCmd.MarkFlagsOneRequired("duplicate-offset", "without-phones")
	deleteCmd.MarkFlagsMutuallyExclusive("duplicate-offset", "without-phones")
}

func deleteContacts(dryRun bool, offset int64, withoutPhones bool) {
	err := storage.PushConfig("")
	if err != nil {
		log.Fatalf("You must have a .env file containg the Dialpad API key")
	}
	defer storage.PopConfig()
	entries, err := contacts.ListContacts("")
	if err != nil {
		log.Fatalf("Download was interrupted: %v", err)
	} else {
		log.Printf("Found %d contacts to clean", len(entries))
	}
	if offset != 0 {
		dupes := contacts.FindOffsetDuplicates(entries, offset)
		log.Printf("Found %d duplicate contacts with offset %d", len(dupes), offset)
		if !dryRun {
			err := contacts.DeleteContacts(dupes)
			if err != nil {
				log.Fatalf("Failed to delete a contact: %v", err)
			} else {
				log.Printf("Deleted %d duplicate contacts", len(dupes))
			}
		}
	}
	if withoutPhones {
		wps := contacts.FindWithoutPhones(entries)
		log.Printf("Found %d contacts without phone numbers", len(wps))
		if !dryRun {
			err := contacts.DeleteContacts(wps)
			if err != nil {
				log.Fatalf("Failed to delete a contact: %v", err)
			} else {
				log.Printf("Deleted %d contacts without phones", len(wps))
			}
		}
	}
}
