/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// eventsCmd represents the events command
var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Operate on webhook events",
	Long: `This command is for manipulating Dialpad webhooks.
You must specify a subcommand to perform an operation.`,
}

func init() {
	rootCmd.AddCommand(eventsCmd)

	eventsCmd.PersistentFlags().StringP("env", "e", "", "processing environment")
}
