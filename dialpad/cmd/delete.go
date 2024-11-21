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
		fromList, _ := cmd.Flags().GetString("from-list")
		wpCount, _ := cmd.Flags().GetCount("without-phones")
		deleteContacts(drCount > 0, fromList, wpCount > 0)
	},
}

func init() {
	contactsCmd.AddCommand(deleteCmd)

	deleteCmd.Args = cobra.ExactArgs(0)
	deleteCmd.Flags().CountP("dry-run", "d", "Don't delete, just report what would be deleted")
	deleteCmd.Flags().String("from-list", "", "Delete duplicates with UIDs offset from master")
	deleteCmd.Flags().Count("without-phones", "Delete contacts that have no phone numbers")
	deleteCmd.MarkFlagsOneRequired("from-list", "without-phones")
	deleteCmd.MarkFlagsMutuallyExclusive("from-list", "without-phones")
}

func deleteContacts(dryRun bool, fromList string, withoutPhones bool) {
	err := storage.PushConfig("")
	if err != nil {
		log.Fatalf("You must have a .env file containg the Dialpad API key")
	}
	defer storage.PopConfig()
	if fromList != "" {
		entries, err := contacts.ImportContacts(fromList)
		if err != nil {
			log.Fatalf("Error importing contacts: %v", err)
		}
		log.Printf("Found %d contacts to delete", len(entries))
		if dryRun {
			log.Printf("Not deleting contacts since dry-run was specified")
			return
		}
		errs := contacts.DeleteContacts(entries)
		if errs != nil {
			log.Printf("Failed to delete %d contacts:", len(errs))
			for _, err := range errs {
				log.Printf("--> %v", err)
			}
		}
		log.Printf("Deleted %d contacts", len(entries)-len(errs))
	}
	if withoutPhones {
		entries, errs := contacts.ListContacts("")
		if errs != nil {
			log.Printf("Dialpad download errors:")
			for _, err := range errs {
				log.Printf("--> %v", err)
			}
			log.Fatalf("Can't continue with an incomplete list of contacts")
		}
		log.Printf("Found %d contacts to inspect", len(entries))
		wps := contacts.FindWithoutPhones(entries)
		log.Printf("Found %d contacts without phone numbers", len(wps))
		if dryRun {
			log.Printf("Not deleting contacts since dry-run was specified")
			return
		}
		errs = contacts.DeleteContacts(wps)
		if errs != nil {
			log.Printf("Failed to delete %d contacts:", len(errs))
			for _, err := range errs {
				log.Printf("--> %v", err)
			}
		}
		log.Printf("Deleted %d contacts without phones", len(entries)-len(errs))
	}
}
