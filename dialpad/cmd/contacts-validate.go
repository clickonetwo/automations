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
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate path_to_csv",
	Short: "Validate contacts",
	Long: `Validate a CSV file of Dialpad contact information.

The first row of the file must be the following header line:
    Creation Date,First_Name,Last_Name,Phones,Email
and each row must have all 5 fields. The processing will print
a list of rows that have errors, and then do an export of
what would be uploaded to Dialpad by the upload command.

The exported CSV uses the same path as the input but with a
".valid.csv" suffix.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Default().SetFlags(0)
		validate(args[0])
	},
}

func init() {
	contactsCmd.AddCommand(validateCmd)

	validateCmd.Args = cobra.ExactArgs(1)
}

func validate(path string) {
	entries, err := contacts.ParseContacts(path, true)
	if err != nil {
		log.Fatalf("Could not read file at path %s: %v", path, err)
	}
	path = strings.TrimSuffix(path, ".csv") + ".valid.csv"
	err = contacts.ExportContacts(entries, path)
	if err != nil {
		log.Fatalf("Could not write file at path %s: %v", path, err)
	}
	log.Printf("%d valid entries written to path %s", len(entries), path)
}
