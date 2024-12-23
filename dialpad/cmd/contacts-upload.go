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

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload [flags] path_to_csv",
	Short: "Upload contacts",
	Long: `Uploads the contacts from a local spreadsheet to Dialpad.
Only new and/or updated contacts are sent.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Default().SetFlags(0)
		count, err := cmd.Flags().GetCount("dry-run")
		if err != nil {
			log.Panic(err)
		}
		upload(args[0], count != 0)
	},
}

func init() {
	contactsCmd.AddCommand(uploadCmd)

	uploadCmd.Args = cobra.ExactArgs(1)
	uploadCmd.Flags().CountP("dry-run", "d", "Don't upload, just report what would be uploaded")
}

func upload(path string, dryRun bool) {
	err := storage.PushConfig("")
	if err != nil {
		log.Fatalf("You must have a .env file containg the Dialpad API key")
	}
	defer storage.PopConfig()
	local, err := contacts.ParseContacts(path, false)
	if err != nil {
		log.Fatalf("Could not read file at path %s: %v", path, err)
	}
	log.Printf("Found %d valid contacts in %s", len(local), path)
	dialpad, errs := contacts.ListContacts("")
	if errs != nil {
		log.Printf("Dialpad download errors:")
		for _, err := range errs {
			log.Printf("--> %v", err)
		}
		log.Fatalf("Can't continue with a incomplete list of contacts")
	}
	log.Printf("Found %d valid contacts in Dialpad", len(dialpad))
	update, create := contacts.DiffEntries(dialpad, local)
	log.Printf("There are %d new contacts and %d updated contacts.", len(create), len(update))
	if dryRun {
		return
	}
	log.Printf("Uploading %d changed contacts to Dialpad...", len(update))
	if errs := contacts.UpdateContacts(update); errs != nil {
		log.Printf("Dialpad update errors:")
		for _, err := range errs {
			log.Printf("--> %v", err)
		}
	}
	log.Printf("Uploading %d new contacts to Dialpad...", len(create))
	if errs := contacts.UpdateContacts(create); errs != nil {
		log.Printf("Dialpad update errors:")
		for _, err := range errs {
			log.Printf("--> %v", err)
		}
	}
	log.Printf("Uploaded all contacts to Dialpad.")
}
