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

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare [flags] path-to-left.csv path-to-right.csv",
	Short: "Compare lists of contacts",
	Long: `Compare two lists of contacts downloaded from Dialpad.
Use a flag to specify the type of comparison. See the flags for details.
The output will vary depending on the type of comparison.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Default().SetFlags(0)
		byId, _ := cmd.Flags().GetCount("by-id")
		byNameAndPhone, _ := cmd.Flags().GetCount("by-name-and-phone")
		if byId > 0 {
			compareById(args[0], args[1])
		} else if byNameAndPhone > 0 {
			compareByNameAndPhone(args[0], args[1])
		} else {
			log.Fatalf("You must specify a type of comparison.")
		}
	},
}

func init() {
	contactsCmd.AddCommand(compareCmd)

	compareCmd.Args = cobra.ExactArgs(2)
	compareCmd.Flags().Count("by-id", "Compare contacts by their unique ID")
	compareCmd.Flags().Count("by-name-and-phone", "Compare contacts by their name and primary phone")
	compareCmd.MarkFlagsOneRequired("by-id", "by-name-and-phone")
	compareCmd.MarkFlagsMutuallyExclusive("by-id", "by-name-and-phone")
}

func compareById(left string, right string) {
	l, err := contacts.ImportContacts(left)
	if err != nil {
		log.Fatalf("Can't import from %q: %v", left, err)
	}
	r, err := contacts.ImportContacts(right)
	if err != nil {
		log.Fatalf("Can't import from %q: %v", right, err)
	}
	both, leftOnly, rightOnly, anomalies := contacts.CompareById(l, r)
	if len(both) > 0 {
		bothPath := strings.TrimSuffix(left, ".csv") + ".both.csv"
		if err := contacts.ExportContacts(both, bothPath); err != nil {
			log.Fatalf("Can't export to %q: %v", bothPath, err)
		}
		log.Printf("%d contacts in both left and right are exported to %q", len(both), bothPath)
	} else {
		log.Printf("There were no contacts in both left and right inputs")
	}
	if len(leftOnly) > 0 {
		leftOnlyPath := strings.TrimSuffix(left, ".csv") + ".only.csv"
		if err := contacts.ExportContacts(leftOnly, leftOnlyPath); err != nil {
			log.Fatalf("Can't export to %q: %v", leftOnlyPath, err)
		}
		log.Printf("%d contacts only in left are exported to %q", len(leftOnly), leftOnlyPath)
	} else {
		log.Printf("There were no contacts that were only in left.")
	}
	if len(rightOnly) > 0 {
		rightOnlyPath := strings.TrimSuffix(right, ".csv") + ".only.csv"
		if err := contacts.ExportContacts(rightOnly, rightOnlyPath); err != nil {
			log.Fatalf("Can't export to %q: %v", rightOnlyPath, err)
		}
		log.Printf("%d contacts only in right are exported to %q", len(rightOnly), rightOnlyPath)
	} else {
		log.Printf("There were no contacts that were only in right.")
	}
	if len(anomalies) > 0 {
		anomaliesPath := strings.TrimSuffix(right, ".csv") + ".anomalies.csv"
		if err := contacts.ExportAnomalies(anomalies, anomaliesPath); err != nil {
			log.Fatalf("Can't export to %q: %v", anomaliesPath, err)
		}
		log.Printf(
			"%d contacts in both but where right content differs from left are exported to %q",
			len(anomalies), anomaliesPath,
		)
	}
}

func compareByNameAndPhone(_ string, _ string) {

}
