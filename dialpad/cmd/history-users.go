/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
	"github.com/clickonetwo/automations/dialpad/internal/users"
)

// usersCmd represents the users command
var usersCmd = &cobra.Command{
	Use:   "users [flags] path-to-users.csv",
	Short: "Manage the users known to the history server",
	Long: `The history server loads a list of users from AWS each time it starts.
This command allows managing that list of users.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		envName, _ := cmd.InheritedFlags().GetString("env")
		if err := storage.PushConfig(envName); err != nil {
			panic(err)
		}
		defer storage.PopConfig()
		update, _ := cmd.Flags().GetCount("update")
		export, _ := cmd.Flags().GetString("export")
		var entries []users.Entry
		if update > 0 {
			entries = updateUsers()
		}
		if export != "" {
			exportUsers(entries, export)
		}
	},
}

func init() {
	historyCmd.AddCommand(usersCmd)
	usersCmd.Args = cobra.NoArgs
	usersCmd.Flags().Count("update", "update users from Dialpad")
	usersCmd.Flags().String("export", "", "export users to path")
	usersCmd.MarkFlagsOneRequired("update", "export")
}

func updateUsers() []users.Entry {
	log.Printf("Fetching users from Dialpad...")
	allUsers, err := users.FetchDialpadUsers()
	if err != nil {
		log.Fatalf("Dialpad fetch failed: %v", err)
	}
	log.Printf("Uploading users to AWS...")
	err = users.UploadUsersList(allUsers)
	if err != nil {
		log.Fatalf("AWS upload failed: %v", err)
	}
	log.Printf("User update complete.")
	return allUsers
}

func exportUsers(entries []users.Entry, path string) {
	if entries == nil {
		log.Printf("Fetching users from AWS...")
		dl, err := users.DownloadUsersList()
		if err != nil {
			log.Fatalf("AWS download failed: %v", err)
		}
		entries = dl
	}
	log.Printf("Exporting %d users in CSV format...", len(entries))
	if err := users.ExportUsers(entries, path); err != nil {
		log.Fatalf("Export failed: %v", err)
	}
	log.Printf("Export of users to %q complete.", path)
}
