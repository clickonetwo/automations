/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// contactsCmd represents the contacts command
var contactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "Operate on dialpad contacts",
	Long: `This command allows you to operate on a spreadsheet of Dialpad contacts.
You must specify an operation and the path to the CSV file(s) to be operated on.`,
}

func init() {
	rootCmd.AddCommand(contactsCmd)
}
