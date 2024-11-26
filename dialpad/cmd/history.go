/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Operate on dialpad history",
	Long: `This command is a parent command for the Dialpad SMS history server.
You must specify one of the subcommands as well as this one.`,
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.PersistentFlags().StringP("env", "e", "", "environment to use")
}
